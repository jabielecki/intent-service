package apisrv

import (
	"database/sql"
	"encoding/json"
	"net/url"
	"sync"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"

	"github.com/tungstenfabric-preview/intent-service/pkg/apisrv/client"
	apicommon "github.com/tungstenfabric-preview/intent-service/pkg/apisrv/common"
	"github.com/tungstenfabric-preview/intent-service/pkg/apisrv/discovery"
	"github.com/tungstenfabric-preview/intent-service/pkg/apisrv/keystone"
	"github.com/tungstenfabric-preview/intent-service/pkg/common"
	"github.com/tungstenfabric-preview/intent-service/pkg/db"
	"github.com/tungstenfabric-preview/intent-service/pkg/db/cache"
	etcdclient "github.com/tungstenfabric-preview/intent-service/pkg/db/etcd"
	"github.com/tungstenfabric-preview/intent-service/pkg/models"
	"github.com/tungstenfabric-preview/intent-service/pkg/services"
	"github.com/tungstenfabric-preview/intent-service/pkg/types"
)

//Server represents Intent API Server.
type Server struct {
	Echo      *echo.Echo
	Keystone  *keystone.Keystone
	dbService *db.Service
	Proxy     *proxyService
	Cache     *cache.DB
}

// NewServer makes a server
func NewServer() (*Server, error) {
	server := &Server{
		Echo: echo.New(),
	}
	return server, nil
}

//SetupService setup service.
//Application with custom logics can embed server struct, and overwrite
//this method.
func (s *Server) SetupService() (services.Service, error) {
	var serviceChain []services.Service

	tv, err := models.NewTypeValidatorWithFormat()
	if err != nil {
		return nil, err
	}

	// ContrailService
	service := &services.ContrailService{
		BaseService:    services.BaseService{},
		TypeValidator:  tv,
		MetadataGetter: s.dbService,
	}

	service.RegisterRESTAPI(s.Echo)
	serviceChain = append(serviceChain, service)

	// ContrailTypeLogicService
	serviceChain = append(serviceChain, &types.ContrailTypeLogicService{
		ReadService:       s.dbService,
		InTransactionDoer: s.dbService,
		AddressManager:    s.dbService,
		IntPoolAllocator:  s.dbService,
		WriteService:      serviceChain[0],
	})
	serviceChain = append(serviceChain, services.NewQuotaCheckerService(s.dbService))

	// EtcdNotifier
	if viper.GetBool("server.notify_etcd") {
		etcdNotifierPath := viper.GetString("etcd.path")
		etcdNotifierService, err := etcdclient.NewNotifierService(etcdNotifierPath)
		if err == nil {
			log.Println("Adding ETCD Notifier Service.")
			serviceChain = append(serviceChain, etcdNotifierService)
		}
	}

	// Put DB Service at the end
	serviceChain = append(serviceChain, s.dbService)

	services.Chain(serviceChain...)

	return service, nil
}

func (s *Server) serveDynamicProxy(endpointStore *apicommon.EndpointStore) {
	s.Proxy = newProxyService(s.Echo, endpointStore, s.dbService)
	s.Proxy.serve()
}

//Init setup the server.
// nolint: gocyclo
func (s *Server) Init() (err error) {
	common.SetLogLevel()

	e := s.Echo
	if viper.GetBool("server.log_api") {
		e.Use(middleware.Logger())
	}
	//e.Use(middleware.Recover())
	//e.Use(middleware.BodyLimit("10M"))

	s.dbService, err = db.NewServiceFromConfig()
	if err != nil {
		return err
	}

	service, err := s.SetupService()
	if err != nil {
		return err
	}

	readTimeout := viper.GetInt("server.read_timeout")
	writeTimeout := viper.GetInt("server.write_timeout")
	e.Server.ReadTimeout = time.Duration(readTimeout) * time.Second
	e.Server.WriteTimeout = time.Duration(writeTimeout) * time.Second

	cors := viper.GetString("server.cors")

	if cors != "" {
		log.Printf("Enabling CORS for %s", cors)
		if cors == "*" {
			log.Printf("cors for * have security issue. DO NOT USE THIS IN PRODUCTION")
		}
		e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:  []string{cors},
			AllowMethods:  []string{echo.GET, echo.PUT, echo.POST, echo.DELETE},
			AllowHeaders:  []string{"X-Auth-Token", "Content-Type"},
			ExposeHeaders: []string{"X-Total-Count"},
		}))
	}

	staticPath := viper.GetStringMapString("server.static_files")
	for prefix, root := range staticPath {
		e.Static(prefix, root)
	}

	proxy := viper.GetStringMapStringSlice("server.proxy")
	for prefix, target := range proxy {
		g := e.Group(prefix)
		g.Use(removePathPrefixMiddleware(prefix))

		t, err := url.Parse(target[0])
		if err != nil {
			return errors.Wrapf(err, "bad proxy target URL: %s", target[0])
		}
		g.Use(proxyMiddleware(t, viper.GetBool("server.proxy.insecure")))
	}
	// serve dynamic proxy based on configured endpoints
	endpointStore := apicommon.MakeEndpointStore() // sync map to store proxy endpoints
	s.serveDynamicProxy(endpointStore)

	keystoneAuthURL := viper.GetString("keystone.authurl")
	var keystoneClient *keystone.KeystoneClient
	if keystoneAuthURL != "" {
		keystoneClient = keystone.NewKeystoneClient(keystoneAuthURL,
			viper.GetBool("keystone.insecure"))
		skipPaths := keystone.GetAuthSkipPaths()
		e.Use(keystone.AuthMiddleware(
			keystoneClient, skipPaths, endpointStore))
	} else if viper.GetString("auth_type") == "no-auth" {
		e.Use(noAuthMiddleware())
	}
	localKeystone := viper.GetBool("keystone.local")
	if localKeystone {
		k, err := keystone.Init(e, endpointStore, keystoneClient)
		if err != nil {
			return errors.Wrap(err, "Failed to init local keystone server")
		}
		s.Keystone = k
	}

	if viper.GetBool("server.enable_grpc") {
		if !viper.GetBool("server.tls.enabled") {
			log.Fatal("GRPC support requires TLS configuraion.")
		}
		log.Debug("enabling grpc")
		var grpcServer *grpc.Server
		if keystoneAuthURL != "" {
			grpcServer = grpc.NewServer(
				grpc.UnaryInterceptor(
					keystone.AuthInterceptor(keystoneClient, endpointStore)))
		} else if viper.GetBool("no_auth") {
			grpcServer = grpc.NewServer(
				grpc.UnaryInterceptor(
					noAuthInterceptor()))
		}
		services.RegisterContrailServiceServer(grpcServer, service)
		e.Use(gRPCMiddleware(grpcServer))
	}

	if viper.GetBool("homepage.enabled") {
		s.setupHomepage()
	}

	s.setupWatchAPI()
	s.setupActionResources()

	if viper.GetBool("recorder.enabled") {
		file := viper.GetString("recorder.file")
		scenario := &TestScenario{
			Workflow: []*Task{},
		}
		var mutex sync.Mutex
		e.Use(middleware.BodyDump(func(c echo.Context, requestBody, responseBody []byte) {
			var data interface{}
			err := json.Unmarshal(requestBody, &data)
			if err != nil {
				log.Debug("malformed json input")
			}
			var expected interface{}
			err = json.Unmarshal(responseBody, &expected)
			if err != nil {
				log.Debug("malformed json response")
			}
			task := &Task{
				Request: &client.Request{
					Method:   c.Request().Method,
					Path:     c.Request().URL.Path,
					Expected: []int{c.Response().Status},
					Data:     data,
				},
				Expect: expected,
			}
			mutex.Lock()
			defer mutex.Unlock()
			scenario.Workflow = append(scenario.Workflow, task)
			err = common.SaveFile(file, scenario)
			if err != nil {
				log.Warn(err)
			}
		}))
	}
	return nil
}

func (s *Server) setupHomepage() {
	dh := discovery.NewHandler(viper.GetString("server.address"))

	services.RegisterSingularPaths(func(path string, name string) {
		dh.Register(path, "", name, "resource-base")
	})
	services.RegisterPluralPaths(func(path string, name string) {
		dh.Register(path, "", name, "collection")
	})

	dh.Register("/fqname-to-id", "POST", "name-to-id", "action")

	// TODO: register sync?

	// TODO action resources
	// TODO documentation
	// TODO VN IP alloc
	// TODO VN IP free
	// TODO subnet IP count
	// TODO set tag
	// TODO security policy draft

	s.Echo.GET("/", dh.Handle)
}

func (s *Server) setupWatchAPI() {
	if !viper.GetBool("cache.enabled") {
		return
	}
	s.Echo.GET("/watch", s.watchHandler)
}

func (s *Server) setupActionResources() {
	s.Echo.POST("/fqname-to-id", s.fqNameToUUIDHandler)
	//TODO handle gRPC
}

// Run runs server.
func (s *Server) Run() error {
	defer func() {
		err := s.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	e := s.Echo
	address := viper.GetString("server.address")
	tlsEnabled := viper.GetBool("server.tls.enabled")
	var keyFile, certFile string
	if tlsEnabled {
		keyFile = viper.GetString("server.tls.key_file")
		certFile = viper.GetString("server.tls.cert_file")

		e.Logger.Fatal(e.StartTLS(address, certFile, keyFile))
	} else {
		e.Logger.Fatal(e.Start(address))
	}
	return nil
}

//Close closes server resources
func (s *Server) Close() error {
	s.Proxy.stop()
	return s.dbService.Close()
}

//DB return db object.
func (s *Server) DB() *sql.DB {
	return s.dbService.DB()
}
