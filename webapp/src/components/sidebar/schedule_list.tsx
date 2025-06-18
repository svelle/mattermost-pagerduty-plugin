// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useState, useRef, useEffect} from 'react';

import type {Schedule} from '@/types/pagerduty';
import type {Theme} from '@/types/theme';

interface Props {
    schedules: Schedule[];
    onScheduleClick: (scheduleId: string) => void;
    theme: Theme;
    loading: boolean;
    error: string | null;
}

const ScheduleList: React.FC<Props> = ({schedules, onScheduleClick, theme, loading, error}) => {
    const [focusedIndex, setFocusedIndex] = useState(-1);
    const scheduleRefs = useRef<Array<HTMLDivElement | null>>([]);

    useEffect(() => {
        // Reset refs array when schedules change
        scheduleRefs.current = scheduleRefs.current.slice(0, schedules.length);
    }, [schedules.length]);

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'ArrowDown') {
            e.preventDefault();
            const nextIndex = focusedIndex < schedules.length - 1 ? focusedIndex + 1 : 0;
            setFocusedIndex(nextIndex);
            scheduleRefs.current[nextIndex]?.focus();
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            const prevIndex = focusedIndex > 0 ? focusedIndex - 1 : schedules.length - 1;
            setFocusedIndex(prevIndex);
            scheduleRefs.current[prevIndex]?.focus();
        } else if (e.key === 'Enter' && focusedIndex >= 0) {
            e.preventDefault();
            onScheduleClick(schedules[focusedIndex].id);
        }
    };

    if (loading) {
        return (
            <div style={{color: theme.centerChannelColor, fontSize: '14px'}}>
                {'Loading schedules...'}
            </div>
        );
    }

    if (error) {
        return (
            <div style={{color: theme.errorTextColor, fontSize: '14px'}}>
                {`Error: ${error}`}
            </div>
        );
    }

    if (schedules.length === 0) {
        return (
            <div style={{color: theme.centerChannelColor, opacity: 0.7, fontSize: '14px'}}>
                {'No schedules found'}
            </div>
        );
    }

    return (
        <div
            className='schedule-list'
            onKeyDown={handleKeyDown}
        >
            <div style={{marginBottom: '12px', color: theme.centerChannelColor, fontSize: '14px', opacity: 0.7}}>
                {`${schedules.length} schedule${schedules.length === 1 ? '' : 's'}`}
            </div>
            {schedules.map((schedule, index) => (
                <div
                    key={schedule.id}
                    ref={(el) => {
                        scheduleRefs.current[index] = el;
                    }}
                    data-testid={`schedule-${schedule.id}`}
                    tabIndex={0}
                    onClick={() => onScheduleClick(schedule.id)}
                    onFocus={() => setFocusedIndex(index)}
                    style={{
                        padding: '12px',
                        marginBottom: '8px',
                        backgroundColor: theme.centerChannelBg,
                        border: `1px solid ${theme.centerChannelColor}20`,
                        borderRadius: '4px',
                        cursor: 'pointer',
                        transition: 'all 0.2s ease',
                        outline: focusedIndex === index ? `2px solid ${theme.buttonBg}` : 'none',
                        outlineOffset: '-1px',
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
