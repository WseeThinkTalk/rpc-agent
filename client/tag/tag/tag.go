package tag

import (
	"context"

	"rpc-agent/client/tag/pb"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	CreateTagRequest     = pb.CreateTagRequest
	CreateTagResponse    = pb.CreateTagResponse
	UpdateTagRequest     = pb.UpdateTagRequest
	UpdateTagResponse    = pb.UpdateTagResponse
	DeleteTagRequest     = pb.DeleteTagRequest
	DeleteTagResponse    = pb.DeleteTagResponse
	TagDetailRequest     = pb.TagDetailRequest
	TagDetailResponse    = pb.TagDetailResponse
	TagListRequest       = pb.TagListRequest
	TagListResponse      = pb.TagListResponse
	HotTagsRequest       = pb.HotTagsRequest
	HotTagsResponse      = pb.HotTagsResponse
	TagResourceRequest   = pb.TagResourceRequest
	TagResourceResponse  = pb.TagResourceResponse
	UntagResourceRequest = pb.UntagResourceRequest
	UntagResourceResponse = pb.UntagResourceResponse
	ResourcesByTagRequest  = pb.ResourcesByTagRequest
	ResourcesByTagResponse = pb.ResourcesByTagResponse
	TagsByResourceRequest  = pb.TagsByResourceRequest
	TagsByResourceResponse = pb.TagsByResourceResponse
	TagItem = pb.TagItem
	ResourceItem = pb.ResourceItem

	Tag interface {
		CreateTag(ctx context.Context, in *CreateTagRequest, opts ...grpc.CallOption) (*CreateTagResponse, error)
		UpdateTag(ctx context.Context, in *UpdateTagRequest, opts ...grpc.CallOption) (*UpdateTagResponse, error)
		DeleteTag(ctx context.Context, in *DeleteTagRequest, opts ...grpc.CallOption) (*DeleteTagResponse, error)
		TagDetail(ctx context.Context, in *TagDetailRequest, opts ...grpc.CallOption) (*TagDetailResponse, error)
		TagList(ctx context.Context, in *TagListRequest, opts ...grpc.CallOption) (*TagListResponse, error)
		HotTags(ctx context.Context, in *HotTagsRequest, opts ...grpc.CallOption) (*HotTagsResponse, error)
		TagResource(ctx context.Context, in *TagResourceRequest, opts ...grpc.CallOption) (*TagResourceResponse, error)
		UntagResource(ctx context.Context, in *UntagResourceRequest, opts ...grpc.CallOption) (*UntagResourceResponse, error)
		ResourcesByTag(ctx context.Context, in *ResourcesByTagRequest, opts ...grpc.CallOption) (*ResourcesByTagResponse, error)
		TagsByResource(ctx context.Context, in *TagsByResourceRequest, opts ...grpc.CallOption) (*TagsByResourceResponse, error)
	}

	defaultTag struct {
		cli zrpc.Client
	}
)

func NewTag(cli zrpc.Client) Tag {
	return &defaultTag{cli: cli}
}

func (m *defaultTag) CreateTag(ctx context.Context, in *CreateTagRequest, opts ...grpc.CallOption) (*CreateTagResponse, error) {
	client := pb.NewTagClient(m.cli.Conn())
	return client.CreateTag(ctx, in, opts...)
}

func (m *defaultTag) UpdateTag(ctx context.Context, in *UpdateTagRequest, opts ...grpc.CallOption) (*UpdateTagResponse, error) {
	client := pb.NewTagClient(m.cli.Conn())
	return client.UpdateTag(ctx, in, opts...)
}

func (m *defaultTag) DeleteTag(ctx context.Context, in *DeleteTagRequest, opts ...grpc.CallOption) (*DeleteTagResponse, error) {
	client := pb.NewTagClient(m.cli.Conn())
	return client.DeleteTag(ctx, in, opts...)
}

func (m *defaultTag) TagDetail(ctx context.Context, in *TagDetailRequest, opts ...grpc.CallOption) (*TagDetailResponse, error) {
	client := pb.NewTagClient(m.cli.Conn())
	return client.TagDetail(ctx, in, opts...)
}

func (m *defaultTag) TagList(ctx context.Context, in *TagListRequest, opts ...grpc.CallOption) (*TagListResponse, error) {
	client := pb.NewTagClient(m.cli.Conn())
	return client.TagList(ctx, in, opts...)
}

func (m *defaultTag) HotTags(ctx context.Context, in *HotTagsRequest, opts ...grpc.CallOption) (*HotTagsResponse, error) {
	client := pb.NewTagClient(m.cli.Conn())
	return client.HotTags(ctx, in, opts...)
}

func (m *defaultTag) TagResource(ctx context.Context, in *TagResourceRequest, opts ...grpc.CallOption) (*TagResourceResponse, error) {
	client := pb.NewTagClient(m.cli.Conn())
	return client.TagResource(ctx, in, opts...)
}

func (m *defaultTag) UntagResource(ctx context.Context, in *UntagResourceRequest, opts ...grpc.CallOption) (*UntagResourceResponse, error) {
	client := pb.NewTagClient(m.cli.Conn())
	return client.UntagResource(ctx, in, opts...)
}

func (m *defaultTag) ResourcesByTag(ctx context.Context, in *ResourcesByTagRequest, opts ...grpc.CallOption) (*ResourcesByTagResponse, error) {
	client := pb.NewTagClient(m.cli.Conn())
	return client.ResourcesByTag(ctx, in, opts...)
}

func (m *defaultTag) TagsByResource(ctx context.Context, in *TagsByResourceRequest, opts ...grpc.CallOption) (*TagsByResourceResponse, error) {
	client := pb.NewTagClient(m.cli.Conn())
	return client.TagsByResource(ctx, in, opts...)
}
