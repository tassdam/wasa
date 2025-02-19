package reqcontext

import (
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

type RequestContext struct {
	ReqUUID uuid.UUID
	Logger  logrus.FieldLogger
}
