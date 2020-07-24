# Pod Contagion
Pod Contagion is an experiment in abusing Kubernetes' container deployment mechanisms. The project tries to create a chain of forever spreading "infection" - a chain spanning multiple generations of pods.

I've used this project as my own intro to Go and as an opportunity to learn some more about Kubernetes. As such the code may be a bit more messy than it needs to be and compatibility with some environments may be limited.

## Concept
A deployment of multiple containers storing the same image is created. The containers are represented as pods in the Kubernetes node. Each container, and thus each pod, is a fundamental unit that can be infected. The software, named Podder, running on each container listens for HTTP requests through which the infection is spread. Additionally, each podder responds to Kubernetes health probes, also through http. An infected podder will start to respond with an error code to Kubernetes which will soon trigger the container's restart. When infected podders start to spread the infection by sending HTTP post requests to other podders (basically sneezing on each other). HTTP communication between podders is enabled by a Kubernetes load balancing service on a common port 80.

The goal is to make an infinite chain of podders getting infected, spreading the infection to others, getting killed by Kubernetes, and getting infected anew after restart by the podders they previously infected.

## The Toy
The software and configuration files are left here for anyone to play with and modify on their own accord. The podder source code in Go is included under `app/` and includes a helper `health` package and a `podder` package which defines much of the podder's functionality. Along with them are two `main` final executable packages: the `generic` (as in, normal, barebones podder) and `starter`. The `starter` can be used to initiate the contagion by starting it's pod outside of the `generic` deployment. Appropriate Dockerfiles producing a minimalistic "from scratch" build are also included.

The images are also hooked to [Docker Hub](https://hub.docker.com/repository/docker/skwiwel/pod-contagion) and are available to be pulled as `skwiwel/pod-contagion:generic` and `skwiwel/pod-contagion:starter`.

In the `kubernetes/` folder included are Kubernetes service, deployment and pod `.yaml` configuration files. Editing these files (`kubernets/deployments/generic.yaml` in particular) is the main way to modify the experiment's flow. 

To run the experiment:
1. Make sure you have a Kubernetes cluster and node up and running.
2. Run `kubectl create -f kubernetes/services/podder.yaml` or create the service some other way. It's important the service is created first so that the podders know where to sneeze (an env variable containing the address is created inside the container).
3. Run `kubectl create -f kubernetes/deployments/generic.yaml` to create the `generic` pod deployments. By default this creates 50 image replicas, but that number may be quite taxing depending on the machine or cloud service. Please adjust the `.yaml` first to make sure the podders can easily be handled.
4. Wait for all the podders to get created and running. Check with `kubectl get pods` for example.
5. Either deploy the `starter` with `kubectl create -f kubernetes/pods/starter.yaml` or send a POST HTTP request with a body of `x-www-form-urlencoded` containing an `action: achoo` field at address `<kubernetes-node-external-IP>:80` (I recommend the [Postman app](https://www.postman.com/) for testing web apps). Some poor podder will reply accordingly if the request is successfully received.
6. Keep checking what is going on with the podders. Unfortunately, I haven't found a simple and good method of gathering logs from a big number of infinitely restarting pods. The easiest way to check the contagion spread is by calling `kubectl get pods` and analyzing the current `RUNNING` and `READY` statuses along with the `RESTARTS` count.

## Experiment Results
Since the principal factor affecting the contagion flow is the Kubernetes create-kill-restart mechanism the results of the experiment are practically nondeterministic - always slightly different. Even so, depending on the deployment configuration the contagion can stop short or thrive possibly forever.

There is one quite big problem I didn't know about before working on this project, though. Kubernetes has an included anti crash-loop mechanism, whose working is indicated by the quite commonly seen container status `CrashLoopBackoff`. The mechanism makes it so that when a container fails it will get restarted with an exponential delay starting at 10 seconds and capping at 300 seconds. Quite unfortunately for this experiment the timer cannot be adjusted as it's hardcoded in the Kubernetes source. Some complaints about it can be found under [kubernetes issue #57291](https://github.com/kubernetes/kubernetes/issues/57291). I can't complain, though, since what I'm trying to do is abuse a mechanism to act in a normally undesirable way and Kubernetes just includes methods preventing such undesirables.

What happens then is if there is no delay between getting infected and showing symptoms (sneezing and kubernetes health status change) the contagion will spread fast but die off quickly after 1 or 2 generations. This is caused by the restart delay timer starting the killed pods after all infected pods were already killed, thus containing the infection. Stopping the contagion is not what we want, though. The podders can be run with their own delay time specifying the time between getting infected and showing symptoms. This can slow the spread enough for the killed pods to be restarted while some podders still are infected. With 50 containers I found it possible to get to ~8 generation before the contagion is contained. And contained it is for one reason - the 300s crash loop backoff delay. Pushing the podders' no symptom timer above 300s can make the contagion go infinitely, but as you may notice, will take a ridiculous amount of time to make any progress.

I haven't found a good solution to that problem. The experiment is certainly not a failure, though. The project explores the workings of the kubernetes mechanism and tests its limits. My intention was for me to learn something new and learn I certainly did. 