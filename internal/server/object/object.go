package object

import "context"

type BucketManager interface {
	MakeBucket(ctx context.Context, bucketName string) (err error)
}
