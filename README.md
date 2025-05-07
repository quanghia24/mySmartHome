# Smart Home Automation System
## Overview
This mySmartHome allows users to:
- Control multiple devices
- Create scheduling operations
- Log system
- Automatically take on control when sensors detect a value above the threshold
- Monitor devices' status with a user-friendly interface
- Statistical system
## Features
1. Device Management
- Control smart devices such as fans, lights, LCD screens, and servos.
- Schedule device operations with predefined timers.
- Configure warning thresholds for sensors to trigger alerts.

2. Sensors & Automation
- Light sensors automatically adjust the brightness according to ambient light.
- Temperature sensors automatically activate fans or air conditioners based on the surrounding temperature.

3. Access Control
- Users can unlock doors by entering a password via their smartphone.

4. Display Interface
- The LCD screen displays temperature, humidity, and device statuses.
- Real-time error notifications and security alerts are provided to users.
- Mobile or web application offers:
  - Full device management
  - Statistics and reports on device usage
  - Real-time notifications on device activity and system warnings

5. Notification system
- Users receive messages when a certain sensor detects a value above a threshold


## System design
![architecture](https://github.com/user-attachments/assets/2a8edd9a-8525-476e-a811-56d9cf41113d)
## Database design
![erd](https://github.com/user-attachments/assets/f1684c37-6e0b-4fa2-b829-b16d3bb4391b)

## System Architecture
- Backend: Manages device control, scheduling, realtime data processing, notification system.
- Frontend: Displays device status, statistics, control options, and alerts.
- Embedded Devices: Handle sensor readings and device actuation based on control signals from the backend.

## Technologies Used
- Backend: Go
- Frontend: ReactNative
- Database: MySQL
- Communication: MQTT / HTTP APIs.
