CLIENT=$(terraform output client_url)  
CLIENT=${CLIENT/\"/}
CLIENT=${CLIENT/\"/}
echo "Waiting for the client to be active"

attempt_counter=0
max_attempts=50

until $(curl --output /dev/null --silent --head --fail $CLIENT); do
    if [ ${attempt_counter} -eq ${max_attempts} ];then
    echo "Max attempts reached"
    exit 1
    fi

    printf '.'
    attempt_counter=$(($attempt_counter+1))
    sleep 5
done

echo "Success, architecture is ready."
echo "To see for yourself, go check out:"
echo $CLIENT