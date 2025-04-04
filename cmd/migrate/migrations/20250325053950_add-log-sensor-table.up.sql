CREATE TABLE IF NOT EXISTS `logs_sensor` (
    `id` INT UNSIGNED AUTO_INCREMENT NOT NULL,
    `type` ENUM('creation', 'data', 'warning') NOT NULL,
    `message` TEXT,
    `sensorId` INT UNSIGNED NOT NULL,
    `userId` INT UNSIGNED NOT NULL,
    `value` VARCHAR(255) NOT NULL,
    `createdAt` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (`id`),
    FOREIGN KEY (`sensorId`) REFERENCES sensors(`feedId`) ON DELETE CASCADE,
    FOREIGN KEY (`userId`) REFERENCES users(`id`) ON DELETE CASCADE
);