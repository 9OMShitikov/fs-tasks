package ext2

const (
	nDirBlocks = 12
	indBlock = nDirBlocks
	dIndBlock = indBlock + 1
	tIndBlock = dIndBlock + 1
	nBlocks = tIndBlock + 1

	fTUnknown	= 0x0
	fTFifo		= 0X1
	fTChrDev	= 0x2
	fTDir		= 0x4
	fTBlkDev	= 0x6
	fTRegFile	= 0x8
	fTSymLink	= 0xA
	fTSock		= 0xC
	fTMax		= 0xF

	rootIno = 2
)