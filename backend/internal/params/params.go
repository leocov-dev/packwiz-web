package params

type Param string

var (
	Token    Param = "token"
	PackSlug Param = "packSlug"
	ModSlug  Param = "modSlug"
	PackId   Param = "packId"
	ModId    Param = "modId"
	ModType  Param = "modType"
	UserID   Param = "userId"
	JobId    Param = "jobId"

	SnapshotId Param = "snapshotId"

	OidcSlug   Param = "slug"
	ProviderId Param = "providerId"
	IdentityId Param = "identityId"
)
