CREATE TABLE IF NOT EXISTS `sensors` (
    `feedId` INT UNSIGNED NOT NULL,
    `feedKey` VARCHAR(255) NOT NULL,
    `title` VARCHAR(255) NOT NULL,
    `type` ENUM('humidity', 'temperature', 'brightness') NOT NULL,
    `userId` INT UNSIGNED NOT NULL,
    `roomId` INT UNSIGNED NOT NULL,
    
    PRIMARY KEY(`feedId`),
    FOREIGN KEY(`userId`) REFERENCES users(`id`) ON DELETE CASCADE,
    FOREIGN KEY(`roomId`) REFERENCES rooms(`id`) ON DELETE CASCADE
)