# Rode Collector SonarQube

[![codecov](https://codecov.io/gh/liatrio/rode-collector-sonarqube/branch/main/graph/badge.svg?token=YK62AO2TNX)](https://codecov.io/gh/liatrio/rode-collector-sonarqube)

[Rode Collector](https://github.com/liatrio/rode-collector-service) for SonarQube 

Collects build metadata from SonarQube to be used to validate automated governance policies. When the collector starts it adds a webhook to SonarQube. It then responds to SonarQube events and sends metadata to a central [Rode Collector](https://github.com/liatrio/rode-collector-service) which stores the metadata in a Grafeas Occurrence.

## Using the Sonarqube Collector
If Sonarqube instance being pointed to is the community edition, an additional step must be followed when executing the sonar scan. This step allows the collector to determine what resource URI should be used.

A command line parameter can be passed in like so, indicating the git url of the project
```
-Dsonar.analysis.resourceUriPrefix=https://github.com/liatrio/springtrader-marketsummary-java
```

It can also be passed into your sonar.properties the same way or your gradle.properties like so:
```
systemProp.sonar.analysis.resourceUriPrefix=https://github.com/liatrio/springtrader-marketsummary-java
```

## Local Setup

```
# Install community edition sonarqube helm chart
helm repo add oteemocharts https://oteemo.github.io/charts
helm install sonarqube oteemocharts/sonarqube
export POD_NAME=$(kubectl get pods --namespace default -l "app=sonarqube,release=sonarqube" -o jsonpath="{.items[0].metadata.name}")\n  echo "Visit http://127.0.0.1:9000 to use your application"\n  kubectl port-forward $POD_NAME 9000:9000 -n default
ngrok http 8080     #needed to setup endpoint for sonarqube

# Scan a project
git clone https://github.com/liatrio/springtrader-marketsummary-java
# Run gradlew with the given command in sonarqube plus the flag mentioned above, ie
./gradlew sonarqube \
  -Dsonar.projectKey=test \
  -Dsonar.host.url=http://127.0.0.1:9000 \
  -Dsonar.login=3c39c7311985d612ac5243bff21347903a143070 \
  -Dsonar.analysis.resourceUriPrefix=https://github.com/liatrio/springtrader-marketsummary-java

# Navigate to the springtrader-marketSummary project and under project settings, set up a webhook url
http://<your-ngrok-url>/webhook/event

# Install Rode via Skaffold
skaffold dev --port-forward
kubectl port-forward svc/rode 50051:50051

# Start SonarQube Collector (eventually replaced via skaffold)
go run .

# Send another scan to sonarqube, and you should see occurrences being received in the rode logs
```
