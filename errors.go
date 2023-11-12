package simver

import (
	"github.com/walteh/terrors"
)

var (
	ErrGit              = terrors.New("simver.ErrGit")
	ErrNoTagsFound      = terrors.New("simver.ErrNoTagsFound")
	ErrFetchingTags     = terrors.New("simver.ErrFetchingTags")
	ErrFindingHead      = terrors.New("simver.ErrFindingHead")
	ErrSettingTag       = terrors.New("simver.ErrSettingTag")
	ErrGettingPRDetails = terrors.New("simver.ErrGettingPRDetails")
	ErrGettingBranch    = terrors.New("simver.ErrGettingBranch")
)
