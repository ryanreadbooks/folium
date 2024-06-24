.PHONY: apiv1
apiv1:
	protoc ./api/v1/*.proto \
		--go_out=api/v1 \
		--go_opt=module=github.com/ryanreadbooks/folium/api/v1 \
		--go-grpc_out=api/v1 \
		--go-grpc_opt=module=github.com/ryanreadbooks/folium/api/v1
