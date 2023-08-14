# Rancher2-kubeconfig

Getting kubeconfig for all available clusters from Rancher2 into one file.  

Set ```RANCHER2_API_TOKEN``` (Bearer token) and ```RANCHER2_API_URL``` (root url like https://myrancher2.domain.com) env vars and run. Default context is first returned cluster.  User will be named myRancher2User, same for all contexts, token is the same as well.  
Configs will be merged into one file, will be written as ```fullkubeconfig``` into working dir.
