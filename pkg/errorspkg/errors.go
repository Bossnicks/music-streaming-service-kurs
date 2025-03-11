package errorspkg

import "errors"

var ErrCommentBanned = errors.New("Администратор запретил вам оставлять комментарии")
