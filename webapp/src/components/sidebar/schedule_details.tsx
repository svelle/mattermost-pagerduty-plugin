// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useMemo} from 'react';

import type {Schedule, ScheduleEntry} from '@/types/pagerduty';
import type {Theme} from '@/types/theme';

interface Props {
    schedule: Schedule;
    theme: Theme;
}

const ScheduleDetails: React.FC<Props> = ({schedule, theme}) => {
    const now = new Date();
    const tomorrow = new Date(now.getTime() + (24 * 60 * 60 * 1000));

    // Get schedule entries for the next 24 hours
    const upcomingEntries = useMemo(() => {
        if (!schedule.final_schedule?.rendered_schedule_entries) {
            return [];
        }

        return schedule.final_schedule.rendered_schedule_entries.
            filter((entry) => {
                const entryStart = new Date(entry.start);
                const entryEnd = new Date(entry.end);

                // Include entries that overlap with our 24-hour window
                return entryEnd > now && entryStart < tomorrow;
            }).
            sort((a, b) => new Date(a.start).getTime() - new Date(b.start).getTime());
    }, [schedule, now, tomorrow]);

    // Find who's currently on call
    const currentOnCall = upcomingEntries.find((entry) => {
        const entryStart = new Date(entry.start);
        const entryEnd = new Date(entry.end);
        return entryStart <= now && entryEnd > now;
    });

    const formatTime = (dateString: string) => {
        const date = new Date(dateString);
        return date.toLocaleTimeString([], {hour: '2-digit', minute: '2-digit'});
    };

    const formatDate = (dateString: string) => {
        const date = new Date(dateString);
        const today = new Date();
        const tomorrow = new Date(today);
        tomorrow.setDate(tomorrow.getDate() + 1);

        if (date.toDateString() === today.toDateString()) {
            return 'Today';
        } else if (date.toDateString() === tomorrow.toDateString()) {
            return 'Tomorrow';
        }
        return date.toLocaleDateString([], {weekday: 'short', month: 'short', day: 'numeric'});
    };

    const getTimeUntilNext = (entry: ScheduleEntry) => {
        const start = new Date(entry.start);
        const diff = start.getTime() - now.getTime();
        const hours = Math.floor(diff / (1000 * 60 * 60));
        const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));

        if (hours > 0) {
            return `in ${hours}h ${minutes}m`;
        }
        return `in ${minutes}m`;
    };

    if (upcomingEntries.length === 0) {
        return (
            <div style={{color: theme.centerChannelColor, opacity: 0.7, fontSize: '14px'}}>
                {'No on-call schedule entries for the next 24 hours'}
            </div>
        );
    }

    return (
        <div className='schedule-details'>
            {currentOnCall && (
                <div
                    style={{
                        backgroundColor: theme.buttonBg,
                        color: theme.buttonColor,
                        padding: '16px',
                        borderRadius: '8px',
                        marginBottom: '20px',
                    }}
                >
                    <div style={{fontSize: '12px', opacity: 0.9, marginBottom: '4px'}}>
                        {'Currently On-Call'}
                    </div>
                    <div style={{fontSize: '18px', fontWeight: 600, marginBottom: '8px'}}>
                        {currentOnCall.user.summary}
                    </div>
                    <div style={{fontSize: '14px', opacity: 0.9}}>
                        {'Until '}{formatTime(currentOnCall.end)}
                        {' ('}{formatDate(currentOnCall.end)}{')'}
                    </div>
                </div>
            )}

            <div style={{marginBottom: '12px'}}>
                <h5 style={{color: theme.centerChannelColor, marginBottom: '12px'}}>
                    {'Next 24 Hours'}
                </h5>
            </div>

            <div className='timeline'>
                {upcomingEntries.map((entry, index) => {
                    const isCurrentlyOnCall = currentOnCall && entry.user.id === currentOnCall.user.id &&
                                            entry.start === currentOnCall.start;
                    const isPastEntry = new Date(entry.end) < now;
                    const isFutureEntry = new Date(entry.start) > now;

                    return (
                        <div
                            key={`${entry.user.id}-${entry.start}-${index}`}
                            style={{
                                display: 'flex',
                                marginBottom: '12px',
                                opacity: isPastEntry ? 0.5 : 1,
                            }}
                        >
                            <div
                                style={{
                                    width: '4px',
                                    backgroundColor: isCurrentlyOnCall ? theme.buttonBg : theme.centerChannelColor,
                                    opacity: isCurrentlyOnCall ? 1 : 0.2,
                                    marginRight: '12px',
                                    borderRadius: '2px',
                                }}
                            />
                            <div style={{flex: 1}}>
                                <div
                                    style={{
                                        padding: '12px',
                                        backgroundColor: isCurrentlyOnCall ? `${theme.buttonBg}15` : theme.centerChannelBg,
                                        border: `1px solid ${isCurrentlyOnCall ? theme.buttonBg : theme.centerChannelColor}20`,
                                        borderRadius: '4px',
                                    }}
                                >
                                    <div style={{display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start'}}>
                                        <div>
                                            <div style={{fontWeight: 500, color: theme.centerChannelColor, marginBottom: '4px'}}>
                                                {entry.user.summary}
                                            </div>
                                            <div style={{fontSize: '13px', color: theme.centerChannelColor, opacity: 0.7}}>
                                                {formatTime(entry.start)}{' - '}{formatTime(entry.end)}
                                            </div>
                                        </div>
                                        <div style={{fontSize: '12px', color: theme.centerChannelColor, opacity: 0.6, textAlign: 'right'}}>
                                            <div>{formatDate(entry.start)}</div>
                                            {isFutureEntry && (
                                                <div style={{marginTop: '2px', fontStyle: 'italic'}}>
                                                    {getTimeUntilNext(entry)}
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    );
                })}
            </div>
        </div>
    );
};

export default ScheduleDetails;

