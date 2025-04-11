# Book Recommendation System

## Launch Instruction

Because we have for today only one of microservices, we made instruction only for `auth-service`. 

- Clone the git repository:

```bash
git clone https://github.com/anuza22/finalProjectGo.git
```

- After that go to the auth-service and use `Makefile` to run docker-compose:

```bash
cd auth-service
make docker-compose-up
```

- You can check the working of this service by going to `/health` endpoint or sending request(p.s. standard port is 8081, you can change this in docker-compose file):

```bash
$ curl localhost:8081/health
> {"service":"auth-service","status":"up","version":"1.0.0"}
```
