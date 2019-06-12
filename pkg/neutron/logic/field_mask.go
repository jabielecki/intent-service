package logic

import "github.com/tungstenfabric-preview/intent-service/pkg/models/basemodels"

func buildDataResourcePath(fields ...string) string {
	fieldPath := basemodels.JoinPath(fields...)
	return basemodels.JoinPath("data.resource", fieldPath)
}

func buildContextPath(fields ...string) string {
	fieldPath := basemodels.JoinPath(fields...)
	return basemodels.JoinPath("context", fieldPath)
}