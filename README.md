# Rancher2-kubeconfig

Getting kubeconfig for all available clusters from Rancher2 into one file.  

Set RANCHER2_API_TOKEN env var and run. Default context is first returned cluster.  
File with config will be written as ```fullkubeconfig``` into working dir. Replace your ```.kube/config``` with it or point your kubectl/k9s etc to it when starting. 
