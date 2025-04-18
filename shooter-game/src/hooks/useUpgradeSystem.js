import { useState } from 'react';

export default function useUpgradeSystem({ networkManager }) {
    const [skillPoints, setSkillPoints] = useState(0);

    const handleUpgrade = (stat) => {
        if (skillPoints <= 0) return;

        networkManager.send({
            type: 'upgrade',
            stat: stat
        });

        setSkillPoints(prev => prev - 1);
    };

    return {
        skillPoints,
        handleUpgrade
    };
}