CREATE TABLE IF NOT EXISTS `schedules` (
    id INT UNSIGNED AUTO_INCREMENT,
    deviceId INT UNSIGNED NOT NULL,
    userId INT UNSIGNED NOT NULL,
    action VARCHAR(50) NOT NULL,
    scheduledTime TIME NOT NULL,            -- for daily schedules
    repeatDays SET('Mon','Tue','Wed','Thu','Fri','Sat','Sun'), -- optional
    timezone VARCHAR(50) DEFAULT 'Asia/Bangkok',     
    isActive BOOLEAN DEFAULT TRUE,
    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY(`id`),
    FOREIGN KEY (`deviceId`) REFERENCES devices(`feedId`) ON DELETE CASCADE,
    FOREIGN KEY (`userId`) REFERENCES users(`id`) ON DELETE CASCADE
);
