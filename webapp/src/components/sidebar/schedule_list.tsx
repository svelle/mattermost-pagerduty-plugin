// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';

import type {Schedule} from '@/types/pagerduty';
import type {Theme} from '@/types/theme';

interface Props {
    schedules: Schedule[];
    onScheduleClick: (scheduleId: string) => void;
    theme: Theme;
}

const ScheduleList: React.FC<Props> = ({schedules, onScheduleClick, theme}) => {
    if (schedules.length === 0) {
        return (
            <div style={{color: theme.centerChannelColor, opacity: 0.7, fontSize: '14px'}}>
                {'No schedules found'}
            </div>
        );
    }

    return (
        <div className='schedule-list'>
            {schedules.map((schedule) => (
                <div
                    key={schedule.id}
                    onClick={() => onScheduleClick(schedule.id)}
                    style={{
                        padding: '12px',
                        marginBottom: '8px',
                        backgroundColor: theme.centerChannelBg,
                        border: `1px solid ${theme.centerChannelColor}20`,
                        borderRadius: '4px',
                        cursor: 'pointer',
                        transition: 'all 0.2s ease',
                    }}
                    onMouseEnter={(e) => {
                        e.currentTarget.style.backgroundColor = `${theme.centerChannelColor}10`;
                        e.currentTarget.style.borderColor = theme.buttonBg;
                    }}
                    onMouseLeave={(e) => {
                        e.currentTarget.style.backgroundColor = theme.centerChannelBg;
                        e.currentTarget.style.borderColor = `${theme.centerChannelColor}20`;
                    }}
                >
                    <div style={{fontWeight: 500, color: theme.centerChannelColor, marginBottom: '4px'}}>
                        {schedule.name}
                    </div>
                    {schedule.description && (
                        <div style={{fontSize: '13px', color: theme.centerChannelColor, opacity: 0.7}}>
                            {schedule.description}
                        </div>
                    )}
                    <div style={{fontSize: '12px', color: theme.centerChannelColor, opacity: 0.5, marginTop: '4px'}}>
                        {schedule.time_zone}
                    </div>
                </div>
            ))}
        </div>
    );
};

export default ScheduleList;
