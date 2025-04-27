# Smart Home Automation System
## Overview

This smart home automation project allows users to control devices, automate operations through sensors, manage access, and monitor device status through a user-friendly application interface.
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


5. Account Management
- Users can log in and log out securely from their accounts.

## System Architecture
- Backend: Manages device control, scheduling, sensor data processing, user authentication, and access control.
- Frontend: Displays device status, statistics, control options, and alerts.
- Embedded Devices: Handle sensor readings and device actuation based on control signals from the backend.

## Technologies Used
- Backend: Go
- Frontend: ReactNative
- Communication: MQTT / HTTP APIs
