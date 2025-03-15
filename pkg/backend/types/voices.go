package types

import (
	"net/http"

	"github.com/moeru-ai/unspeech/pkg/apierrors"
	"github.com/samber/mo"
)

type VoicesRequestOptions struct {
	Backend string `json:"provider"`
}

func NewVoicesRequestOptions(request *http.Request) mo.Result[VoicesRequestOptions] {
	provider := request.URL.Query().Get("provider")
	if provider == "" {
		return mo.Err[VoicesRequestOptions](
			apierrors.
				NewErrInvalidArgument().
				WithDetail("provider is required").
				WithSourceParameter("provider"),
		)
	}

	return mo.Ok(VoicesRequestOptions{
		Backend: provider,
	})
}
