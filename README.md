# Rode Collector SonarQube

[Rode Collector](https://github.com/liatrio/rode-collector-service) for SonarQube 

Collects build metadata from SonarQube to be used to validate automated governance policies. When the collector starts it adds a webhook to SonarQube. It then responds to SonarQube events and sends metadata to a central [Rode Collector](https://github.com/liatrio/rode-collector-service) which stores the metadata in a Grafeas Occurrence.

## Arguments
| Argument | Environment Variables | Description | Default | Required |
|----------|-----------------------|-------------|---------|----------|
| -url            | URL            | Collector URL for SonarQube to send events to |  | [x] |
| -sonar-url      | SONAR_URL      | SonarQube URL | http://localhost:9000 | [x] |
| -sonar-username | SONAR_USERNAME | Username to authenticate with SonarQube | admin | if token not set |
| -sonar-password | SONAR_PASSWORD | Password to authenticate with SonarQube | admin | if token not set |
| -sonar-token    | SONAR_TOKEN    | Token to authenticate with SonarQube | | overrides username and password |
