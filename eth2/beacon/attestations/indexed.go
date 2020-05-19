package attestations

import (
	"errors"
	"fmt"
	. "github.com/protolambda/zrnt/eth2/core"
	"github.com/protolambda/zrnt/eth2/meta"
	"github.com/protolambda/zrnt/eth2/util/bls"
	"github.com/protolambda/zrnt/eth2/util/ssz"
	"sort"
)

type CommitteeIndices []ValidatorIndex

func (ci *CommitteeIndices) Limit() uint64 {
	return MAX_VALIDATORS_PER_COMMITTEE
}

type IndexedAttestation struct {
	AttestingIndices CommitteeIndices
	Data             AttestationData
	Signature        BLSSignature
}

type AttestationValidator interface {
	meta.RegistrySize
	meta.Pubkeys
	meta.Versioning
}

// Verify validity of slashable_attestation fields.
func (indexedAttestation *IndexedAttestation) Validate(m AttestationValidator) error {
	// wrap it in validator-sets. Does not sort it, but does make checking if it is a lot easier.
	indices := ValidatorSet(indexedAttestation.AttestingIndices)

	// Verify max number of indices
	if count := len(indices); count > MAX_VALIDATORS_PER_COMMITTEE {
		return fmt.Errorf("invalid indices count in indexed attestation: %d", count)
	}

	// The indices must be sorted
	if !sort.IsSorted(indices) {
		return errors.New("attestation indices are not sorted")
	}

	// Verify if the indices are unique. Simple O(n) check, since they are already sorted.
	for i := 1; i < len(indices); i++ {
		if indices[i-1] == indices[i] {
			return fmt.Errorf("attestation indices at %d and %d are duplicate, both: %d", i-1, i, indices[i])
		}
	}

	// Check the last item of the sorted list to be a valid index,
	// if this one is valid, the others are as well, since they are lower.
	if len(indices) > 0 && !m.IsValidIndex(indices[len(indices)-1]) {
		return errors.New("attestation indices contains out of range index")
	}

	pubkeys := make([]BLSPubkey, 0, 2)
	for _, i := range indices {
		pubkeys = append(pubkeys, m.Pubkey(i))
	}

	// empty attestation
	if len(pubkeys) <= 0 {
		return errors.New("in phase 0 no empty attestation signatures are allowed")
	}

	if !bls.FastAggregateVerify(pubkeys,
		ComputeSigningRoot(
			ssz.HashTreeRoot(&indexedAttestation.Data, AttestationDataSSZ),
			m.GetDomain(DOMAIN_BEACON_ATTESTER, indexedAttestation.Data.Target.Epoch)),
		indexedAttestation.Signature,
	) {
		return errors.New("could not verify BLS signature for indexed attestation")
	}

	return nil
}
