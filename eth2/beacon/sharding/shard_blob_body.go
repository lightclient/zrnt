package sharding

import (
	"github.com/protolambda/zrnt/eth2/beacon/common"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

const POINTS_PER_SAMPLE = 8

type ShardData []common.BLSPoint

func (d *ShardData) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return tree.ReadRootsLimited(dr, (*[]common.Root)(d), POINTS_PER_SAMPLE*spec.MAX_SAMPLES_PER_BLOCK)
}

func (d ShardData) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return tree.WriteRoots(w, d)
}

func (d ShardData) ByteLength(spec *common.Spec) (out uint64) {
	return uint64(len(d)) * 32
}

func (d *ShardData) FixedLength(spec *common.Spec) uint64 {
	return 0 // it's d list, no fixed length
}

func (d ShardData) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	length := uint64(len(d))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &d[i]
		}
		return nil
	}, length, POINTS_PER_SAMPLE*spec.MAX_SAMPLES_PER_BLOCK)
}

func ShardDataType(spec *common.Spec) ListTypeDef {
	return ListType(common.BLSPointType, POINTS_PER_SAMPLE*spec.MAX_SAMPLES_PER_BLOCK)
}

func ShardBlobBodyType(spec *common.Spec) *ContainerTypeDef {
	return ContainerType("ShardBlobBody", []FieldDef{
		{"commitment", DataCommitmentType},
		{"degree_proof", BLSCommitmentType},
		{"data", ShardDataType(spec)},
		{"max_priority_fee_per_sample", common.GweiType},
		{"max_fee_per_sample", common.GweiType},
	})
}

type ShardBlobBody struct {
	// The actual data commitment
	Commitment DataCommitment `json:"commitment" yaml:"commitment"`
	// Proof that the degree < commitment.length
	DegreeProof BLSCommitment `json:"degree_proof" yaml:"degree_proof"`
	// The actual data. Should match the commitment and degree proof.
	Data ShardData `json:"data" yaml:"data"`
	// fee payment fields (EIP 1559 like)
	MaxPriorityFeePerSample common.Gwei `json:"max_priority_fee_per_sample" yaml:"max_priority_fee_per_sample"`
	MaxFeePerSample         common.Gwei `json:"max_fee_per_sample" yaml:"max_fee_per_sample"`
}

func (b *ShardBlobBody) Deserialize(spec *common.Spec, dr *codec.DecodingReader) error {
	return dr.FixedLenContainer(&b.Commitment, &b.DegreeProof, spec.Wrap(&b.Data), &b.MaxPriorityFeePerSample, &b.MaxFeePerSample)
}

func (b *ShardBlobBody) Serialize(spec *common.Spec, w *codec.EncodingWriter) error {
	return w.FixedLenContainer(&b.Commitment, &b.DegreeProof, spec.Wrap(&b.Data), &b.MaxPriorityFeePerSample, &b.MaxFeePerSample)
}

func (b *ShardBlobBody) ByteLength(spec *common.Spec) uint64 {
	return codec.ContainerLength(&b.Commitment, &b.DegreeProof, spec.Wrap(&b.Data), &b.MaxPriorityFeePerSample, &b.MaxFeePerSample)
}

func (b *ShardBlobBody) FixedLength(spec *common.Spec) uint64 {
	// dynamic size due to data List
	return 0
}

func (b *ShardBlobBody) HashTreeRoot(spec *common.Spec, hFn tree.HashFn) common.Root {
	return hFn.HashTreeRoot(&b.Commitment, &b.DegreeProof, spec.Wrap(&b.Data), &b.MaxPriorityFeePerSample, &b.MaxFeePerSample)
}
