package magic

import "bytes"

type (
	Detector func(raw []byte, limit uint32) bool
)

var (
	// Bz2 matches the bzip2 file format
	Bz2 = prefix([]byte{0x42, 0x5A, 0x68})

	// Gz matches the gzip file format
	Gz = prefix([]byte{0x1f, 0x8b})

	// Rar1 and Rar2 match the rar file format
	Rar1 = prefix([]byte{0x52, 0x61, 0x72, 0x21, 0x1a, 0x07, 0x00})
	Rar2 = prefix([]byte{0x52, 0x61, 0x72, 0x21, 0x1a, 0x07, 0x01, 0x00})
	Rar  = func(raw []byte, limit uint32) bool {
		return Rar1(raw, limit) || Rar2(raw, limit)
	}

	// Tar1 and Tar2 match the tar file format
	Tar1 = offset([]byte{0x75, 0x73, 0x74, 0x61, 0x72, 0x00, 0x30, 0x30}, 257)
	Tar2 = offset([]byte{0x75, 0x73, 0x74, 0x61, 0x72, 0x20, 0x20, 0x00}, 257)
	// Keeping the below here for reference. Can be used instead of the above
	// Tar  = offset([]byte("ustar"), 257)

	// Zip1, Zip2, and Zip3 match the zip file format
	Zip1 = prefix([]byte{0x50, 0x4b, 0x03, 0x04})
	Zip2 = prefix([]byte{0x50, 0x4b, 0x05, 0x06})
	Zip3 = prefix([]byte{0x50, 0x4b, 0x07, 0x08})
	Zip  = func(raw []byte, limit uint32) bool {
		return Zip1(raw, limit) || Zip2(raw, limit) || Zip3(raw, limit)
	}

	// SevenZ matches the 7z file format
	SevenZ = prefix([]byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C})
)

func Tar(raw []byte, limit uint32) bool {
	return Tar1(raw, limit) || Tar2(raw, limit)

}

func prefix(sigs ...[]byte) Detector {
	return func(raw []byte, limit uint32) bool {
		for _, sig := range sigs {
			if bytes.HasPrefix(raw, sig) {
				return true
			}
		}
		return false
	}
}

func offset(sig []byte, offset int) Detector {
	return func(raw []byte, limit uint32) bool {
		return len(raw) > offset && bytes.HasPrefix(raw[offset:], sig)
	}
}
