package protocol

// Pure-Go GF(2^8) Cauchy Reed-Solomon Forward Error Correction.
//
// Legacy Lagvex FEC used single-parity XOR (recovers at most 1 lost packet per
// block). Real residential Wi-Fi, 4G/5G and congested subsea routes suffer burst
// losses (2-3 consecutive packets dropped at 5-15% loss). Systematic Cauchy
// Reed-Solomon coding produces M parity shards per K data shards and recovers
// any L <= M simultaneous drops inside a block with zero retransmission RTT.

const (
	// FECMaxDataShards bounds K in RS(K, M) blocks.
	FECMaxDataShards = 16
	// FECMaxParityShards bounds M in RS(K, M) blocks.
	FECMaxParityShards = 6
)

var (
	gfExp [512]byte
	gfLog [256]byte
)

func init() {
	// Primitive polynomial for GF(2^8): x^8 + x^4 + x^3 + x^2 + 1 (0x11d).
	x := 1
	for i := 0; i < 255; i++ {
		gfExp[i] = byte(x)
		gfLog[x] = byte(i)
		x <<= 1
		if x&0x100 != 0 {
			x ^= 0x11d
		}
	}
	for i := 255; i < 512; i++ {
		gfExp[i] = gfExp[i-255]
	}
}

func gfMul(a, b byte) byte {
	if a == 0 || b == 0 {
		return 0
	}
	return gfExp[int(gfLog[a])+int(gfLog[b])]
}

func gfInv(a byte) byte {
	if a == 0 {
		return 0
	}
	return gfExp[255-int(gfLog[a])]
}

// gfMulAddSlice computes dst[i] ^= coeff * src[i] over GF(2^8).
func gfMulAddSlice(dst, src []byte, coeff byte) {
	if coeff == 0 {
		return
	}
	n := len(src)
	if len(dst) < n {
		n = len(dst)
	}
	if coeff == 1 {
		for i := 0; i < n; i++ {
			dst[i] ^= src[i]
		}
		return
	}
	logC := int(gfLog[coeff])
	for i := 0; i < n; i++ {
		if s := src[i]; s != 0 {
			dst[i] ^= gfExp[int(gfLog[s])+logC]
		}
	}
}

// cauchyCoeff returns element (parityIdx, dataIdx) of a systematic Cauchy
// encoding matrix over GF(2^8). Because x = k + parityIdx and y = dataIdx are
// strictly disjoint for all k + m <= 256, x ^ y is never zero and every square
// submatrix is guaranteed invertible.
func cauchyCoeff(k, parityIdx, dataIdx int) byte {
	return gfInv(byte(k+parityIdx) ^ byte(dataIdx))
}

// rsEncodeBlock generates m Cauchy Reed-Solomon parity shards for k data shards.
// Shards of differing lengths are implicitly zero-padded to the max length in
// the block (original lengths are carried in FECPayload.Lengths).
func rsEncodeBlock(shards [][]byte, m int) [][]byte {
	k := len(shards)
	if k == 0 || m <= 0 {
		return nil
	}
	maxLen := 0
	for _, s := range shards {
		if len(s) > maxLen {
			maxLen = len(s)
		}
	}
	if maxLen == 0 {
		return nil
	}

	parities := make([][]byte, m)
	for r := 0; r < m; r++ {
		p := make([]byte, maxLen)
		for c := 0; c < k; c++ {
			gfMulAddSlice(p, shards[c], cauchyCoeff(k, r, c))
		}
		parities[r] = p
	}
	return parities
}

// gfInvertMatrix inverts an n x n matrix over GF(2^8) via Gauss-Jordan
// elimination. Returns nil, false if singular.
func gfInvertMatrix(mat [][]byte) ([][]byte, bool) {
	n := len(mat)
	aug := make([][]byte, n)
	for i := 0; i < n; i++ {
		row := make([]byte, 2*n)
		copy(row[:n], mat[i])
		row[n+i] = 1
		aug[i] = row
	}

	for col := 0; col < n; col++ {
		pivot := -1
		for r := col; r < n; r++ {
			if aug[r][col] != 0 {
				pivot = r
				break
			}
		}
		if pivot < 0 {
			return nil, false
		}
		if pivot != col {
			aug[col], aug[pivot] = aug[pivot], aug[col]
		}

		invDiag := gfInv(aug[col][col])
		if invDiag != 1 {
			for c := col; c < 2*n; c++ {
				aug[col][c] = gfMul(aug[col][c], invDiag)
			}
		}

		for r := 0; r < n; r++ {
			if r == col {
				continue
			}
			factor := aug[r][col]
			if factor == 0 {
				continue
			}
			for c := col; c < 2*n; c++ {
				aug[r][c] ^= gfMul(factor, aug[col][c])
			}
		}
	}

	inv := make([][]byte, n)
	for i := 0; i < n; i++ {
		inv[i] = append([]byte(nil), aug[i][n:]...)
	}
	return inv, true
}

// rsRecoverMissing reconstructs L missing data shards using L received parity
// shards (where L = len(missingIndices) <= len(parityIndices)).
func rsRecoverMissing(
	k int,
	present map[int][]byte,
	missingIndices []int,
	missingLens []int,
	parityIndices []int,
	parityData [][]byte,
) ([][]byte, bool) {
	l := len(missingIndices)
	if l == 0 || len(parityIndices) < l || len(parityData) < l {
		return nil, false
	}

	maxLen := 0
	for i := 0; i < l; i++ {
		if len(parityData[i]) > maxLen {
			maxLen = len(parityData[i])
		}
		if missingLens[i] <= 0 || missingLens[i] > len(parityData[i]) {
			return nil, false
		}
	}

	// Build L x L submatrix A and RHS vectors (parity minus known data contributions).
	mat := make([][]byte, l)
	rhs := make([][]byte, l)
	for i := 0; i < l; i++ {
		pIdx := parityIndices[i]
		row := make([]byte, l)
		for j := 0; j < l; j++ {
			row[j] = cauchyCoeff(k, pIdx, missingIndices[j])
		}
		mat[i] = row

		vec := make([]byte, maxLen)
		copy(vec, parityData[i])
		for c, shard := range present {
			gfMulAddSlice(vec, shard, cauchyCoeff(k, pIdx, c))
		}
		rhs[i] = vec
	}

	inv, ok := gfInvertMatrix(mat)
	if !ok {
		return nil, false
	}

	recovered := make([][]byte, l)
	for j := 0; j < l; j++ {
		out := make([]byte, missingLens[j])
		for i := 0; i < l; i++ {
			gfMulAddSlice(out, rhs[i][:missingLens[j]], inv[j][i])
		}
		recovered[j] = out
	}
	return recovered, true
}
