package signer_app

import (
	"encoding/json"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
)

func (s *SignerApp) getAggregationProof(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestHash := common.HexToHash(r.URL.Query().Get("requestHash"))
	proof, err := s.cfg.Repo.GetAggregationProof(ctx, requestHash)
	if err != nil {
		handleError(ctx, w, err)
		return
	}

	type response struct {
		VerificationType uint32 `json:"verification_type"`
		MessageHash      []byte `json:"message_hash"`
		Proof            []byte `json:"proof"`
	}

	resp := response{
		VerificationType: uint32(proof.VerificationType),
		MessageHash:      proof.MessageHash,
		Proof:            proof.Proof,
	}

	body, err := json.Marshal(resp)
	if err != nil {
		handleError(ctx, w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
