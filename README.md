# ccloud-admin

Run this command to delete all topics in the cluuster with the given environment name prefix. The program executes as a DRY-RUN by default and only deletes topics if the --force option is used
```console
./build/linux/amd64/ccloud-delete \
-RESTEndpoint https://zz-endpoint.us-east-1.aws.confluent.cloud:443 \
-clusterID zz-cluster \
-environment dev \
-username MyUsErNaMe \
-password MyPaSsWoRd
# --force
```
