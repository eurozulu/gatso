# Build the image
docker build -t todolist ..
if [ $? -eq 0 ]
then
# Tag the image
docker tag todolist eurospoofer/todolist:0.1.0

docker login

docker push eurospoofer/todolist:0.1.0
fi

