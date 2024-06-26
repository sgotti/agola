// Copyright 2019 Sorint.lab
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/sorintlab/errors"

	"agola.io/agola/internal/services/gateway/action"
	"agola.io/agola/internal/util"
	csapitypes "agola.io/agola/services/configstore/api/types"
	cstypes "agola.io/agola/services/configstore/types"
	gwapitypes "agola.io/agola/services/gateway/api/types"
)

func createSecretResponse(s *csapitypes.Secret) *gwapitypes.SecretResponse {
	return &gwapitypes.SecretResponse{
		ID:         s.ID,
		Name:       s.Name,
		ParentPath: s.ParentPath,
	}
}

type SecretsHandler struct {
	log zerolog.Logger
	ah  *action.ActionHandler
}

func NewSecretsHandler(log zerolog.Logger, ah *action.ActionHandler) *SecretsHandler {
	return &SecretsHandler{log: log, ah: ah}
}

func (h *SecretsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h.do(r)
	if util.HTTPError(w, err) {
		h.log.Err(err).Send()
		return
	}

	if err := util.HTTPResponse(w, http.StatusOK, res); err != nil {
		h.log.Err(err).Send()
	}
}

func (h *SecretsHandler) do(r *http.Request) ([]*gwapitypes.SecretResponse, error) {
	ctx := r.Context()
	query := r.URL.Query()
	_, tree := query["tree"]
	_, removeoverridden := query["removeoverridden"]

	parentType, parentRef, err := GetConfigTypeRef(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	areq := &action.GetSecretsRequest{
		ParentType:       parentType,
		ParentRef:        parentRef,
		Tree:             tree,
		RemoveOverridden: removeoverridden,
	}
	cssecrets, err := h.ah.GetSecrets(ctx, areq)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	secrets := make([]*gwapitypes.SecretResponse, len(cssecrets))
	for i, s := range cssecrets {
		secrets[i] = createSecretResponse(s)
	}

	return secrets, nil
}

type CreateSecretHandler struct {
	log zerolog.Logger
	ah  *action.ActionHandler
}

func NewCreateSecretHandler(log zerolog.Logger, ah *action.ActionHandler) *CreateSecretHandler {
	return &CreateSecretHandler{log: log, ah: ah}
}

func (h *CreateSecretHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h.do(r)
	if util.HTTPError(w, err) {
		h.log.Err(err).Send()
		return
	}

	if err := util.HTTPResponse(w, http.StatusCreated, res); err != nil {
		h.log.Err(err).Send()
	}
}

func (h *CreateSecretHandler) do(r *http.Request) (*gwapitypes.SecretResponse, error) {
	ctx := r.Context()
	parentType, parentRef, err := GetConfigTypeRef(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var req gwapitypes.CreateSecretRequest
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&req); err != nil {
		return nil, util.NewAPIErrorWrap(util.ErrBadRequest, err)
	}

	areq := &action.CreateSecretRequest{
		Name:             req.Name,
		ParentType:       parentType,
		ParentRef:        parentRef,
		Type:             cstypes.SecretType(req.Type),
		Data:             req.Data,
		SecretProviderID: req.SecretProviderID,
		Path:             req.Path,
	}
	cssecret, err := h.ah.CreateSecret(ctx, areq)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	res := createSecretResponse(cssecret)

	return res, nil
}

type UpdateSecretHandler struct {
	log zerolog.Logger
	ah  *action.ActionHandler
}

func NewUpdateSecretHandler(log zerolog.Logger, ah *action.ActionHandler) *UpdateSecretHandler {
	return &UpdateSecretHandler{log: log, ah: ah}
}

func (h *UpdateSecretHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	res, err := h.do(r)
	if util.HTTPError(w, err) {
		h.log.Err(err).Send()
		return
	}

	if err := util.HTTPResponse(w, http.StatusOK, res); err != nil {
		h.log.Err(err).Send()
	}
}

func (h *UpdateSecretHandler) do(r *http.Request) (*gwapitypes.SecretResponse, error) {
	ctx := r.Context()
	vars := mux.Vars(r)
	secretName := vars["secretname"]

	parentType, parentRef, err := GetConfigTypeRef(r)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var req gwapitypes.UpdateSecretRequest
	d := json.NewDecoder(r.Body)
	if err := d.Decode(&req); err != nil {
		return nil, util.NewAPIErrorWrap(util.ErrBadRequest, err)
	}
	areq := &action.UpdateSecretRequest{
		SecretName: secretName,

		Name:             req.Name,
		ParentType:       parentType,
		ParentRef:        parentRef,
		Type:             cstypes.SecretType(req.Type),
		Data:             req.Data,
		SecretProviderID: req.SecretProviderID,
		Path:             req.Path,
	}
	cssecret, err := h.ah.UpdateSecret(ctx, areq)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	res := createSecretResponse(cssecret)

	return res, nil
}

type DeleteSecretHandler struct {
	log zerolog.Logger
	ah  *action.ActionHandler
}

func NewDeleteSecretHandler(log zerolog.Logger, ah *action.ActionHandler) *DeleteSecretHandler {
	return &DeleteSecretHandler{log: log, ah: ah}
}

func (h *DeleteSecretHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.do(r)
	if util.HTTPError(w, err) {
		h.log.Err(err).Send()
		return
	}

	if err := util.HTTPResponse(w, http.StatusNoContent, nil); err != nil {
		h.log.Err(err).Send()
	}
}

func (h *DeleteSecretHandler) do(r *http.Request) error {
	ctx := r.Context()
	vars := mux.Vars(r)
	secretName := vars["secretname"]

	parentType, parentRef, err := GetConfigTypeRef(r)
	if err != nil {
		return errors.WithStack(err)
	}

	err = h.ah.DeleteSecret(ctx, parentType, parentRef, secretName)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
