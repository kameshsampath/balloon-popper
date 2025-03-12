/*
 * Copyright (c) 2025.  Kamesh Sampath <kamesh.sampath@hotmail.com>
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 *
 */

class BalloonGame {
    constructor(character, playerName) {
        console.log("Initializing game for:", playerName, "as", character);
        this.character = character;
        this.playerName = playerName;
        this.score = 0;
        this.bonusHits = 0;
        this.regularHits = 0;
        this.isActive = true;

        // Canvas setup
        this.canvas = document.getElementById("gameCanvas");
        this.ctx = this.canvas.getContext("2d");
        this.resizeCanvas();

        // Game state
        this.balloons = [];
        this.popEffects = [];
        this.lastSpawnTime = 0;
        this.spawnInterval = 2000;
        this.balloonRadius = 30;
        this.hitAreaMultiplier = 1.5;

        // Initialize WebSocket
        this.connectWebSocket();

        // Set up event listeners
        window.addEventListener("resize", () => this.resizeCanvas());
        this.canvas.addEventListener("click", (e) => this.handleClick(e));

        // Game Config
        fetch("/config")
            .then((response) => response.json())
            .then((config) => {
                this.gameConfig = config;
                console.log("Loaded game config:", config);
                // Start game loop
                this.gameLoop();
            })
            .catch((error) =>
                console.error("Failed to load game config:", error)
            );
    }

    resizeCanvas() {
        const container = this.canvas.parentElement;
        const rect = container.getBoundingClientRect();
        this.canvas.width = rect.width;
        this.canvas.height = 500;
        console.log(
            "Canvas resized to:",
            this.canvas.width,
            this.canvas.height
        );
    }

    connectWebSocket() {
        console.log("Connecting WebSocket for player:", this.playerName);
        this.ws = new WebSocket(
            `ws://${window.location.host}/ws/${this.playerName}`
        );

        this.ws.onopen = () => {
            console.log("WebSocket connected");
        };

        this.ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            console.log("Received WebSocket message:", data);
            if (data.type === "score_update") {
                this.updateScore(data.event);
            }
        };
    }

    createBalloon() {
        console.trace("Creating Balloon with Config", this.gameConfig);
        // Get character's favorite colors from game config
        const favoriteColors =
            this.gameConfig.character_favorites[this.character];
        const bonusProbability = this.gameConfig.bonus_probability;

        console.trace("Colors:", Object.keys(this.gameConfig.colors));
        // Available colors for regular balloons
        const regularColors = Object.keys(this.gameConfig.colors).filter(
            (color) => !favoriteColors.includes(color)
        ); // Remove favorite colors

        // Determine if this will be a bonus balloon
        const isBonus = Math.random() < bonusProbability;

        // Select color based on bonus status
        const color = isBonus
            ? favoriteColors[Math.floor(Math.random() * favoriteColors.length)]
            : regularColors[Math.floor(Math.random() * regularColors.length)];

        const balloon = {
            x: Math.random() * (this.canvas.width - 60) + 30,
            y: this.canvas.height + 30,
            radius: this.balloonRadius,
            color: color,
            speed: Math.random() * 1 + 1,
            bobOffset: 0,
            bobSpeed: Math.random() * 0.05 + 0.02,
            bobTime: Math.random() * Math.PI * 2,
            scale: 1,
            isBonus: isBonus, // Track bonus status
            sparkleAngle: 0, // For bonus balloon effect
        };

        this.balloons.push(balloon);
    }

    createPopEffect(x, y, color) {
        const particles = [];
        const particleCount = 8;

        for (let i = 0; i < particleCount; i++) {
            const angle = (i / particleCount) * Math.PI * 2;
            particles.push({
                x,
                y,
                color,
                speed: 5,
                angle,
                life: 1,
            });
        }

        this.popEffects.push({
            particles,
            age: 0,
        });
    }

    updateBalloons() {
        for (let i = this.balloons.length - 1; i >= 0; i--) {
            const balloon = this.balloons[i];
            balloon.y -= balloon.speed;
            balloon.bobTime += balloon.bobSpeed;
            balloon.bobOffset = Math.sin(balloon.bobTime) * 5;

            if (balloon.y < -50) {
                this.balloons.splice(i, 1);
            }
        }

        // Update pop effects
        for (let i = this.popEffects.length - 1; i >= 0; i--) {
            const effect = this.popEffects[i];
            effect.age += 0.05;

            effect.particles.forEach((particle) => {
                particle.x += Math.cos(particle.angle) * particle.speed;
                particle.y += Math.sin(particle.angle) * particle.speed;
                particle.life -= 0.05;
            });

            if (effect.age > 1) {
                this.popEffects.splice(i, 1);
            }
        }
    }

    handleClick(event) {
        if (!this.isActive) return;

        const rect = this.canvas.getBoundingClientRect();
        const x =
            (event.clientX - rect.left) * (this.canvas.width / rect.width);
        const y =
            (event.clientY - rect.top) * (this.canvas.height / rect.height);

        for (let i = this.balloons.length - 1; i >= 0; i--) {
            const balloon = this.balloons[i];
            const dx = x - balloon.x;
            const dy = y - (balloon.y + balloon.bobOffset);
            const distance = Math.sqrt(dx * dx + dy * dy);

            if (distance < balloon.radius * this.hitAreaMultiplier) {
                // Create pop effect
                this.createPopEffect(
                    balloon.x,
                    balloon.y + balloon.bobOffset,
                    balloon.color
                );

                // Send pop event to server
                if (this.ws && this.ws.readyState === WebSocket.OPEN) {
                    this.ws.send(
                        JSON.stringify({
                            balloon_color: balloon.color,
                            event_ts: new Date().toISOString(),
                            player: this.playerName,
                            character: this.character,
                        })
                    );
                }

                this.balloons.splice(i, 1);
                break;
            }
        }
    }

    drawBalloon(balloon) {
        const ctx = this.ctx;
        const x = balloon.x;
        const y = balloon.y + balloon.bobOffset;

        // Draw string
        ctx.beginPath();
        ctx.moveTo(x, y + balloon.radius);
        ctx.lineTo(x, y + balloon.radius + 20);
        ctx.strokeStyle = "#666";
        ctx.lineWidth = 2;
        ctx.stroke();

        // Draw balloon
        ctx.beginPath();
        ctx.arc(x, y, balloon.radius, 0, Math.PI * 2);
        ctx.fillStyle = balloon.color;
        ctx.fill();

        // Add highlight
        const gradient = ctx.createRadialGradient(
            x - balloon.radius / 3,
            y - balloon.radius / 3,
            balloon.radius / 10,
            x,
            y,
            balloon.radius
        );
        gradient.addColorStop(0, "rgba(255, 255, 255, 0.4)");
        gradient.addColorStop(1, "rgba(255, 255, 255, 0)");
        ctx.fillStyle = gradient;
        ctx.fill();

        // Add sparkle effect for bonus balloons
        if (balloon.isBonus) {
            // Update sparkle rotation
            balloon.sparkleAngle += 0.05;

            // Draw sparkle
            const sparklePoints = 8;
            const outerRadius = balloon.radius * 1.3;
            const innerRadius = balloon.radius * 1.1;

            ctx.beginPath();
            for (let i = 0; i < sparklePoints * 2; i++) {
                const radius = i % 2 === 0 ? outerRadius : innerRadius;
                const angle =
                    (i * Math.PI) / sparklePoints + balloon.sparkleAngle;
                const sparkleX = x + Math.cos(angle) * radius;
                const sparkleY = y + Math.sin(angle) * radius;

                if (i === 0) {
                    ctx.moveTo(sparkleX, sparkleY);
                } else {
                    ctx.lineTo(sparkleX, sparkleY);
                }
            }
            ctx.closePath();
            ctx.strokeStyle = "gold";
            ctx.lineWidth = 2;
            ctx.stroke();
        }
    }

    drawPopEffects() {
        const ctx = this.ctx;

        this.popEffects.forEach((effect) => {
            effect.particles.forEach((particle) => {
                if (particle.life > 0) {
                    ctx.beginPath();
                    ctx.arc(particle.x, particle.y, 3, 0, Math.PI * 2);
                    ctx.fillStyle = particle.color;
                    ctx.globalAlpha = particle.life;
                    ctx.fill();
                    ctx.globalAlpha = 1;
                }
            });
        });
    }

    draw() {
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
        this.balloons.forEach((balloon) => this.drawBalloon(balloon));
        this.drawPopEffects();
    }

    updateScore(eventData) {
        console.log("Updating score:", eventData);
        this.score += eventData.score;
        if (eventData.favorite_color_bonus) {
            this.bonusHits++;
        } else {
            this.regularHits++;
        }

        // Update UI
        const scoreElement = document.getElementById("score");
        const bonusElement = document.getElementById("bonusHits");
        const regularElement = document.getElementById("regularHits");

        if (scoreElement) scoreElement.textContent = `Score: ${this.score}`;
        if (bonusElement)
            bonusElement.textContent = `Bonus Hits: ${this.bonusHits}`;
        if (regularElement)
            regularElement.textContent = `Regular Hits: ${this.regularHits}`;

        // Add bonus animation
        if (eventData.favorite_color_bonus) {
            scoreElement.classList.add("bonus");
            setTimeout(() => scoreElement.classList.remove("bonus"), 1000);
        }
    }

    stopGame() {
        this.isActive = false;
        this.balloons = [];
        this.popEffects = [];
    }

    gameLoop() {
        if (this.isActive) {
            const currentTime = Date.now();
            if (currentTime - this.lastSpawnTime > this.spawnInterval) {
                this.createBalloon();
                this.lastSpawnTime = currentTime;
            }

            this.updateBalloons();
        }

        this.draw();
        requestAnimationFrame(() => this.gameLoop());
    }
}
