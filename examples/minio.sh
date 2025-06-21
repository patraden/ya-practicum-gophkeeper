# redis notifications setup and validation
# https://min.io/docs/minio/linux/administration/monitoring/publish-events-to-redis.html

# create alias:
mc alias set local https://localhost:9000 gophkeeper gophkeeper --insecure
# check alias created:
mc alias list
# list buckets
mc ls local --insecure
# get redis notify config:
mc admin config get local notify_redis
# apply config:
mc admin config set local notify_redis \
  enable=on \
  format=namespace \
  address=redis:6379 \
  key=minioevents \
  region=eu-central-1
# restart minio deployment:
mc admin service restart local
# get ARN resource:
mc admin info --json local
# confirm event notifications setup:
mc event list local/mysecondbucket arn:minio:sqs:eu-central-1:gophkeeper:redis --insecure

# put object on presigned url
curl --cacert ./deployments/.certs/ca.cert -X PUT --upload-file testfile.txt \
  "https://localhost:9000/myfirstdbucket/testfile.txt?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=gophkeeper%2F20250611%2Feu-central-1%2Fs3%2Faws4_request&X-Amz-Date=20250611T200037Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&X-Amz-Signature=b1533bc629171471e229ad8e2f5059322c6d2bc9017f58df7da7a90839d5a391"


mc idp openid add local identity_openid vendor="keycloak" keycloak_realm="master" scopes="openid,email,preferred_username" client_id="minio" client_secret="u79g3gjmfg0IjrGPLFggbT9muRDdyp8S" config_url="http://localhost:8080/realms/master/.well-known/openid-configuration" redirect_uri_dynamic="on" display_name="SSO_IDENTIFIER" --insecure

mc idp openid add local identity_openid \                                                                                                          
vendor="keycloak" \ OpenID IDP server configuration                                                                                                         
keycloak_realm="master" \nID IDP server configuration                                                                                                        
scopes="minio-authorization" \                                                                                                                                
client_id="minio" \                                                                                                                                           
client_secret="u79g3gjmfg0IjrGPLFggbT9muRDdyp8S" \tion folder (default: "/root/.mc") [$MC_CONFIG_DIR]                                                         
config_url="http://keycloak:8080/realms/master/.well-known/openid-configuration" \                                                                            
redirect_uri_dynamic="on"


mc idp openid info local PRIMARY --insecure