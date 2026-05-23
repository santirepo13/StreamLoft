#StreamLoft 

this is for redirecting streams on several clients, i have the gui for windows, deploy on a db and the api and media server run on another server


how to deploy:

1. install srs 
apt update
apt install -y wget unzip curl

cd /opt
wget -O SRS-CentOS7-x86_64-6.0-r0.zip https://github.com/ossrs/srs/releases/download/v6.0-r0/SRS-CentOS7-x86_64-6.0-r0.zip
unzip SRS-CentOS7-x86_64-6.0-r0.zip

config file location
/opt/SRS-CentOS7-x86_64-6.0-r0/usr/local/srs/conf/srs.conf

nano conf/srs.conf

srs.conf :
``
# SRS Media Server Configuration for StreamLoft

listen              1935;
max_connections     1000;
daemon              on;
pid                 ./objs/srs.pid;
srs_log_file        ./objs/srs.log;

http_api {
    enabled         on;
    listen          1985;
}

http_server {
    enabled         on;
    listen          8081;
    dir             ./objs/nginx/html;
}

vhost __defaultVhost__ {
    http_hooks {
        enabled         on;
        on_publish      http://172.86.73.79:8080/stream/start;
        on_unpublish    http://172.86.73.79:8080/stream/stop;
    }

    http_remux {
        enabled         on;
        mount           /live/[stream].flv;
    }

    min_latency     on;
    mw_latency      1000;
    chunk_size      4096;
    queue_length    200;
}

``

to start srs 
./objs/srs -c conf/srs.conf

Check running
ps aux | grep -i srs

check ports
ss -lntp | grep srs

2. Deploy the API via building it:

Prerequisites on the API server (172.86.73.79):

apt update
apt install -y golang git postgresql-client

Clone the repository:


cd /opt/StreamLoft/backend-api

Create the environment file from the example:

cp .env.example /opt/StreamLoft/api.env
nano /opt/StreamLoft/api.env

api.env:
``
DATABASE_HOST=87.239.135.39
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=L...
DATABASE_NAME=streamloft
API_PORT=8080
SRS_URL=http://172.86.73.79:8081
RTMP_URL=rtmp://172.86.73.79/live
JWT_SECRET=ncTuR7x67SX1Kl6PIWIH1eYeQT46gLOTEN1TOonMsQM=
LOG_LEVEL=info
ENCRYPTION_KEY_PATH=/opt/StreamLoft/streamloft_master.key
``

Build the binary:

go mod download
go build -o streamloft-api ./cmd/streamloft-api

Run the API:

./streamloft-api

Or run it as a background service with systemd:

nano /etc/systemd/system/streamloft-api.service

streamloft-api.service:
``
[Unit]
Description=StreamLoft API
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/StreamLoft/backend-api
ExecStart=/opt/StreamLoft/backend-api/streamloft-api
Restart=always
RestartSec=10
Environment="ENV_FILE=/opt/StreamLoft/api.env"

[Install]
WantedBy=multi-user.target
``

Enable and start:

systemctl daemon-reload
systemctl enable streamloft-api
systemctl start streamloft-api

Check running

ps aux | grep streamloft

check ports

ss -lntp | grep 8080

Check status:

systemctl status streamloft-api

Check logs:

journalctl -u streamloft-api -f

3. Set up the database:

On the database server (87.239.135.39), install PostgreSQL if not already installed:

apt update
apt install -y postgresql postgresql-contrib

Start PostgreSQL:

systemctl start postgresql
systemctl enable postgresql

Create the database and user:

sudo -u postgres psql

Inside psql:
``
CREATE DATABASE streamloft;
CREATE USER streamloft_user WITH PASSWORD 'L...';
GRANT ALL PRIVILEGES ON DATABASE streamloft TO streamloft_user;
\q
``

Run the schema:

psql -h localhost -U streamloft_user -d streamloft -f /opt/StreamLoft/database/schema.sql

Verify tables were created:

psql -h localhost -U streamloft_user -d streamloft -c "\dt"

4. Configure the encryption key:

Generate a master encryption key on the API server:

openssl rand -hex 32 > /opt/StreamLoft/streamloft_master.key

Set permissions:

chmod 600 /opt/StreamLoft/streamloft_master.key

this one is sketchy because sometimes i run this cmd and it creates a 34 bit one , must check correct 32 bit created key

5. Deploy the Windows app:

Prerequisites on the Windows machine:

- .NET 10 Runtime or .NET 10 SDK
- Windows 10 or later

Build from source:

cd windows-app/StreamLoftApp
dotnet build --configuration Release

Publish as a self-contained executable:

dotnet publish -c Release -r win-x64 --self-contained true -o ./publish

The published files will be in windows-app/StreamLoftApp/publish/

Distribute the publish folder to client machines. Run the app:

StreamLoftApp.exe

Or create a desktop shortcut to StreamLoftApp.exe for end users.

6. Verify the deployment:

Check SRS is running:

curl http://172.86.73.79:8081/

Check API health:

curl http://172.86.73.79:8080/health

Expected response: {"status":"ok"}

Check the Windows app can reach the API:

Open the app and confirm the login screen loads and can authenticate.

7. Admin database operations:

Admin SQL scripts are located in database/admin-sql/

Open pgAdmin on the database server (87.239.135.39) and run the scripts as needed
