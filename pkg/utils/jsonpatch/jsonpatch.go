package jsonpatch

import (
	"encoding/json"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/moeru-ai/unspeech/pkg/utils"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

type JSONPatchOperation string

const (
	JSONPatchOperationAdd     JSONPatchOperation = "add"
	JSONPatchOperationRemove  JSONPatchOperation = "remove"
	JSONPatchOperationReplace JSONPatchOperation = "replace"
)

type JSONPatchOperationObject struct {
	Operation JSONPatchOperation `json:"op"`
	Path      string             `json:"path"`
	Value     any                `json:"value,omitempty"`
}

func NewPatches(operations ...mo.Option[JSONPatchOperationObject]) []byte {
	return lo.Must(json.Marshal(utils.MapOptionsPresent(operations)))
}

func NewReplace(path string, to any) mo.Option[JSONPatchOperationObject] {
	return mo.Some(JSONPatchOperationObject{
		Operation: JSONPatchOperationReplace,
		Path:      path,
		Value:     to,
	})
}

func NewAdd(path string, value any) mo.Option[JSONPatchOperationObject] {
	return mo.Some(JSONPatchOperationObject{
		Operation: JSONPatchOperationAdd,
		Path:      path,
		Value:     value,
	})
}

func NewRemove(path string) mo.Option[JSONPatchOperationObject] {
	return mo.Some(JSONPatchOperationObject{
		Operation: JSONPatchOperationRemove,
		Path:      path,
	})
}

type ApplyOptions jsonpatch.ApplyOptions

func ApplyPatches(bytes []byte, applyOpt mo.Option[ApplyOptions], patches ...mo.Option[JSONPatchOperationObject]) mo.Result[[]byte] {
	patch, err := jsonpatch.DecodePatch(NewPatches(patches...))
	if err != nil {
		return mo.Err[[]byte](err)
	}

	patched, err := patch.Apply(bytes)
	if err != nil {
		return mo.Err[[]byte](err)
	}

	return mo.Ok[[]byte](patched)
}
