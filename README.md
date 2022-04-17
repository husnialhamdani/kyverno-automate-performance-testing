# kyverno-automate-performance-testing

### Setup cluster, metrics server, install Kyverno & Policies
$ bash setup.sh

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

#### Config on Github Actions pipeline

```yaml
jobs:
  setup-cluster:
    runs-on: ubuntu-latest
    steps:
      # Checks-out your repository under $GITHUB_WORKSPACE, so your job can access it
      - uses: actions/checkout@v3
      
      - name: Run a multi-line script
        run: |
          ls
          
      - name: setup cluster
        uses: appleboy/ssh-action@master
        with:
          host: ${{ secrets.HOST }}
          username: ${{ secrets.USERNAME }}
          key: ${{ secrets.KEY }}
          passphrase: ${{ secrets.PASSPHRASE }}
          port: ${{ secrets.PORT }}
          script: |
            ls
            git clone https://github.com/husnialhamdani/kyverno-automate-performance-testing.git dir
            bash ~/dir/setup.sh
   
     # - name: storing report
     #   uses: actions/upload-artifact@v3
     #   with:
     #     name: report
     #     path: report.png

  automation:
    runs-on: ubuntu-latest
    timeout-minutes: 40
    steps:
      - uses: actions/checkout@v3
      
      - name: run automation
        uses: appleboy/ssh-action@master
        env:
          EMAILFROM: ${{ secrets.EMAILFROM }}
          EMAILPASS: ${{ secrets.EMAILPASS }}
          EMAILTO: ${{ secrets.EMAILTO }}
        with:
          host: ${{ secrets.HOST }}
          username: ${{ secrets.USERNAME }}
          key: ${{ secrets.KEY }}
          passphrase: ${{ secrets.PASSPHRASE }}
          port: ${{ secrets.PORT }}
          script: |
            cd ~/dir
            /usr/local/go/bin/go mod download
            /usr/local/go/bin/go run main.go -scales=large
```
