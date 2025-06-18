// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';

import type {Schedule} from '@/types/pagerduty';
import type {Theme} from '@/types/theme';

interface Props {
    schedule: Schedule | null;
    onBack: () => void;
    theme: Theme;
    loading: boolean;
}

const ScheduleDetails: React.FC<Props> = ({schedule, onBack, theme, loading}) => {
    if (loading) {
        return (
            <div style={{padding: '20px', color: theme.centerChannelColor}}>
                {'Loading schedule details...'}
            </div>
        );
    }

    if (!schedule) {
        return (
            <div style={{padding: '20px', color: theme.centerChannelColor}}>
                {'No schedule selected'}
            </div>
        );
    }

    const entries = schedule.final_schedule?.rendered_schedule_entries || [];

    return (
        <div style={{padding: '20px'}}>
            <div style={{marginBottom: '16px'}}>
                <button
                    onClick={onBack}
                    style={{
                        backgroundColor: 'transparent',
                        border: 'none',
                        color: theme.linkColor,
                        cursor: 'pointer',
                        padding: '4px 8px',
                        fontSize: '14px',
                    }}
                >
                    {'← Back'}
                </button>
            </div>

            <div style={{marginBottom: '16px'}}>
                <h3 style={{color: theme.centerChannelColor, margin: 0}}>
                    {schedule.name}
                </h3>
                {schedule.time_zone && (
                    <div style={{color: theme.centerChannelColor, opacity: 0.7, fontSize: '14px', marginTop: '4px'}}>
                        {schedule.time_zone}
                    </div>
                )}
            </div>

            <div style={{marginBottom: '16px'}}>
                <h4 style={{color: theme.centerChannelColor, marginBottom: '12px'}}>
                    {'On-Call Schedule'}
                </h4>

                {!schedule.final_schedule && (
                    <div style={{color: theme.centerChannelColor, opacity: 0.7, fontSize: '14px'}}>
                        {'No on-call schedule available'}
                    </div>
                )}

                {schedule.final_schedule && entries.length === 0 && (
                    <div style={{color: theme.centerChannelColor, opacity: 0.7, fontSize: '14px'}}>
                        {'No on-call entries for this schedule'}
                    </div>
                )}

                {entries.map((entry, index) => (
                    <div
                        key={`${entry.user.id}-${entry.start}-${index}`}
                        data-testid={`schedule-entry-${index}`}
                        style={{
                            padding: '12px',
                            backgroundColor: theme.centerChannelBg,
                            border: `1px solid ${theme.centerChannelColor}20`,
                            borderRadius: '4px',
                            marginBottom: '8px',
                            display: 'flex',
                            alignItems: 'center',
                        }}
                    >
                        {entry.user.avatar_url && (
                            <img
                                src={entry.user.avatar_url}
                                alt={entry.user.name}
                                style={{
                                    width: '32px',
                                    height: '32px',
                                    borderRadius: '50%',
                                    marginRight: '12px',
                                }}
                            />
                        )}
                        <div style={{flex: 1}}>
                            <div style={{fontWeight: 500, color: theme.centerChannelColor}}>
                                {entry.user.name}
                            </div>
                            <div style={{fontSize: '12px', color: theme.centerChannelColor, opacity: 0.7}}>
                                {entry.user.email}
                            </div>
                            <div style={{fontSize: '12px', color: theme.centerChannelColor, opacity: 0.5, marginTop: '4px'}}>
                                {new Date(entry.start).toLocaleString()}{' - '}{new Date(entry.end).toLocaleString()}
                            </div>
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};

export default ScheduleDetails;
