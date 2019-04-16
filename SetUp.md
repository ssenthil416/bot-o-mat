
// Bot-O-Mat

//Use to get clone to copy thid code to your local path
>> git clone https://github.com/RedVentures22/bot-o-mat-ssenthil416


//Build

>> go build -o bot *.go

>> Modify inputParam.yaml, based on your reqirement

//Service File : Modify serivce and Copy service to service location
>> cp ./misc/botomat.service  /usr/lib/systemd/system/botomat.service

//Start service
>>sudo systemctl daemon-reload
>>sudo systemctl start botomat.service

//Stop Serice 
>>sudo systemctl stop botomat.service

//Status Serice 
>>sudo systemctl status botomat.service

//Also added Webserver to show Robot status
>>After running thr app, pls check this URL "http://localhost:3222/status"



Note :

> All logs are logged in the path /tmp/botomat.log
> Very five minute Robot status is displayed
> Finally status is also displayed.
> Not handle : robot will not start from were it stopped for some reason.
> Try to added json viewer extension to chrome browser for better look
