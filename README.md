# godnslog

A dns&amp;http log server for verify SSRF/XXE/RFI/RCE vulnerability 

## build frontend

requirements: 

`yarn`

```
cd frontend
yarn install
yarn build
```
	
## build backend

requirements: 

`golang >= 1.13.0`

```bash
go build
```

## docker build

```bash
docker build -t "user/godnslog" .
```

For Chinese user:

```bash
docker build -t "user/godnslog" -f DockerfileCN .
