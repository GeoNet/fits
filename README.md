# FITS

Field Time Series data.

### Compilation

There are scripts `build.sh` and `build-push.sh` for building Docker containers.

### Database

There is a Docker file which can be used to create a DB image with the DB schema ready to use:

```
docker build --rm=true -t 615890063537.dkr.ecr.ap-southeast-2.amazonaws.com/fits-db:12 -f etc/Dockerfile database
docker push 615890063537.dkr.ecr.ap-southeast-2.amazonaws.com/fits-db:12
```

To start the database:

```
docker run -e POSTGRES_PASSWORD={yourpassword} -p 5432:5432  615890063537.dkr.ecr.ap-southeast-2.amazonaws.com/fits-db:12
```

Add test data to the DB with:

```
./etc/scripts/initdb-test.sh postgres {yourpassword}
```

Full DB init and load a small amount of test data with:

```
cd scripts; ./initdb.sh postgres {yourpassword}
```

#### Logical Model

The database logical model.

![database logical model](etc/ddl/FITS_Logical_Model.png)


### Deployment

Deployment on AWS Elastic Beanstalk (EB) using Docker containers.

#### fits-api

There are files for EB - both to deploy the application and also set
up logging from the container (application) to CloudWatch Logs.  Create a zip file and then upload the 
zip to EB.

```
cd deploy
zip fits.zip Dockerrun.aws.json .ebextensions/*
```
