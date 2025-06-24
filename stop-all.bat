@echo off
echo Stopping all Kafka UI processes...
taskkill /F /IM go.exe /T >nul 2>&1
taskkill /F /IM node.exe /T >nul 2>&1
echo All processes stopped. 