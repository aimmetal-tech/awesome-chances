package model

import "errors"

var ErrFeedbackConflict = errors.New("相同事件 ID 已用于不同反馈")
