package params

type Param string

var (
	Token    Param = "token"
	PackSlug Param = "packSlug"
	ModSlug  Param = "modSlug"
	PackId   Param = "packId"
	ModId    Param = "modId"
	ModType  Param = "modType"

	// InstanceFile is the instance zip's file name; any value is accepted.
	InstanceFile Param = "instanceFile"
	UserID       Param = "userId"
	JobId        Param = "jobId"

	SnapshotId Param = "snapshotId"

	OidcSlug   Param = "slug"
	ProviderId Param = "providerId"
	IdentityId Param = "identityId"
)

// InstanceZipDir is the path segment, under a consumer pack, that the instance
// zip is served from as InstanceZipDir/:InstanceFile.
const InstanceZipDir = "multimc"
