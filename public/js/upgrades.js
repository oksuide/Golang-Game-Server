export class UpgradeSystem {
    constructor(game) {
        this.game = game;
        this.menu = document.getElementById('upgrade-menu');
        this.skillPoints = 0;
        this.upgradeMenuVisible = false;
        this.socket = game.network.socket; // Добавляем доступ к сокету
    }

    init() {
        document.addEventListener('keydown', (e) => {
            if (e.key === 'u' && this.skillPoints > 0) {
                this.toggleMenu(!this.upgradeMenuVisible);
            }
        });

        document.addEventListener('click', (e) => this.handleUpgradeClick(e));
    }

    toggleMenu(show) {
        this.menu.style.display = show ? 'block' : 'none';
        this.upgradeMenuVisible = show;

        if (show) {
            document.getElementById('skill-points').textContent = `Очков: ${this.skillPoints}`;
            // Блокируем кнопки если очков нет
            document.querySelectorAll('.upgrade-btn').forEach(btn => {
                btn.disabled = this.skillPoints <= 0;
            });
        }
    }

    handleUpgradeClick(e) {
        if (!e.target.classList.contains('upgrade-btn')) return;
        if (this.skillPoints <= 0) return;

        const btn = e.target;
        const stat = btn.dataset.stat;

        console.log('[Upgrade] Кнопка нажата:', stat);
        console.log('[Upgrade] Отправка на сервер:', { upgradeStat: stat });

        try {
            this.socket.send(JSON.stringify({
                type: 'upgrade',
                stat: stat
            }));

            this.skillPoints--;
            document.getElementById('skill-points').textContent = `Очков: ${this.skillPoints}`;
            btn.disabled = this.skillPoints <= 0;

            if (this.skillPoints <= 0) {
                console.log('[Upgrade] Скрытие меню улучшений');
                this.toggleMenu(false);
            }
        } catch (err) {
            console.error('[Upgrade] Ошибка отправки:', err);
        }
    }
}