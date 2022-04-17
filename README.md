# kyverno-automate-performance-testing

### Setup cluster, metrics server, snstall Kyverno & Policies
$ bash automation.sh

### Automate Performance Testing
#### Scales options
"xs": 100
"small": 500
"medium": 1000
"large": 2000
"xl": 3000

#### Example commands
$ go mod download

$ go run main.go -scales=large

