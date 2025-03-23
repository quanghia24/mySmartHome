CREATE TABLE IF NOT EXISTS `logs` (
    `id` INT UNSIGNED AUTO_INCREMENT NOT NULL,
    `type` ENUM('creation', 'onoff', 'schedule', 'warning') NOT NULL,
    `message` TEXT,
    `deviceId` INT UNSIGNED NOT NULL,
    `userId` INT UNSIGNED NOT NULL,
    `value` VARCHAR(255) NOT NULL,
    `createdAt` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (`id`),
    FOREIGN KEY (`deviceId`) REFERENCES devices(`feedId`),
    FOREIGN KEY (`userId`) REFERENCES users(`id`)
);