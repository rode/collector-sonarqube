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