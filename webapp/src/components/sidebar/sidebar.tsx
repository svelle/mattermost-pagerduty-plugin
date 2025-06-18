// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React, {useEffect, useState} from 'react';

import ScheduleDetails from './schedule_details';
import ScheduleList from './schedule_list';

import client from '@/client/client';
import type {Schedule} from '@/types/pagerduty';
import type {Theme} from '@/types/theme';

interface Props {
    theme: Theme;
}

const PagerDutySidebar: React.FC<Props> = ({theme}) => {
    const [schedules, setSchedules] = useState<Schedule[]>([]);
    const [selectedSchedule, setSelectedSchedule] = useState<Schedule | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [loadingDetails, setLoadingDetails] = useState(false);

    useEffect(() => {
        fetchSchedules();
    }, []);

    const fetchSchedules = async () => {
        try {
            setLoading(true);
            setError(null);

            const schedulesData = await client.getSchedules();
            setSchedules(schedulesData.schedules || []);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load schedules');
        } finally {
            setLoading(false);
        }
    };

    const handleScheduleClick = async (scheduleId: string) => {
        // If clicking the same schedule, go back to list view
        if (selectedSchedule?.id === scheduleId) {
            setSelectedSchedule(null);
            return;
        }

        setLoadingDetails(true);
        try {
            const scheduleDetails = await client.getScheduleDetails(scheduleId);
            setSelectedSchedule(scheduleDetails);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load schedule details');
        } finally {
            setLoadingDetails(false);
        }
    };

    const handleBack = () => {
        setSelectedSchedule(null);
    };

    const handleRefresh = () => {
        if (selectedSchedule) {
            handleScheduleClick(selectedSchedule.id);
        } else {
            fetchSchedules();
        }
    };

    return (
        <div
            className='pagerduty-sidebar'
            style={{height: '100%', display: 'flex', flexDirection: 'column'}}
        >
            <div
                style={{
                    padding: '16px',
                    borderBottom: `1px solid ${theme.centerChannelColor}20`,
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                }}
            >
                <div style={{display: 'flex', alignItems: 'center', gap: '12px'}}>
                    {selectedSchedule && (
                        <button
                            onClick={handleBack}
                            style={{
                                backgroundColor: 'transparent',
                                color: theme.linkColor,
                                border: 'none',
                                padding: '4px',
                                cursor: 'pointer',
                                fontSize: '18px',
                                lineHeight: 1,
                            }}
                            title='Back to schedules'
                        >
                            {'←'}
                        </button>
                    )}
                    <h3 style={{margin: 0, color: theme.centerChannelColor}}>
                        {selectedSchedule && selectedSchedule.name ? selectedSchedule.name : 'PagerDuty Schedules'}
                    </h3>
                </div>
                <button
                    onClick={handleRefresh}
                    style={{
                        backgroundColor: 'transparent',
                        color: theme.linkColor,
                        border: 'none',
                        padding: '4px 8px',
                        cursor: 'pointer',
                        fontSize: '14px',
                    }}
                >
                    {'Refresh'}
                </button>
            </div>

            <div style={{flex: 1, overflow: 'auto', padding: '16px'}}>
                {selectedSchedule || loadingDetails ? (
                    <ScheduleDetails
                        schedule={selectedSchedule}
                        onBack={handleBack}
                        theme={theme}
                        loading={loadingDetails}
                    />
                ) : (
                    <ScheduleList
                        schedules={schedules}
                        onScheduleClick={handleScheduleClick}
                        theme={theme}
                        loading={loading}
                        error={error}
                    />
                )}
            </div>
        </div>
    );
};

export default PagerDutySidebar;