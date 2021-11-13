package ext2

type superBlock struct {
	INodeCount    			uint32
	BlockCount  			uint32
	RBlockCount 			uint32
	FreeBlockCount		uint32
	FreeInodeCount			uint32
	FirstDataBlock			uint32
	LogBlockSize			uint32
	LogClusterSize			uint32
	BlockPerGroup			uint32
	ClusterPerGroup			uint32
	InodePerGroup			uint32
	MTime					uint32
	WTime					uint32
	MntCount				uint16
	MaxMntCount				uint16
	Magic					uint16
	State					uint16
	Errors					uint16
	MinorRevLevel			uint16
	LastCheck				uint32
	CheckInterval			uint32
	CreatorOS				uint32
	RevLevel				uint32
	DefResUID				uint16
	DefResGID				uint16

	FirstIno				uint32
	INodeSize				uint16
	BlockGroupNr			uint16
	FeatureCompat			uint32
	FeatureInCompat			uint32
	FeatureRoCompat			uint32
	Uuid					[16]byte
	VolumeName				[16]byte
	LastMounted				[64]byte
	AlgorithmUsageBitmap	uint32
};
