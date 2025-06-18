// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import React from 'react';

import PagerDutySidebar from './sidebar';

import client from '@/client/client';
import {render, screen, waitFor, fireEvent, mockTheme} from '@/test-utils';

// Mock the client module
jest.mock('@/client/client');
const mockClient = client as jest.Mocked<typeof client>;

// Mock child components
jest.mock('./schedule_list', () => ({
    __esModule: true,
    default: ({schedules, onScheduleClick, loading, error}: any) => (
        <div data-testid='schedule-list'>
            {loading && <div>{'Loading schedules...'}</div>}
            {error && <div>{'Error: '}{error}</div>}
            {schedules.map((schedule: any) => (
                <button
                    key={schedule.id}
                    data-testid={`schedule-${schedule.id}`}
                    onClick={() => onScheduleClick(schedule.id)}
                >
                    {schedule.name}
                </button>
            ))}
        </div>
    ),
}));

jest.mock('./schedule_details', () => ({
    __esModule: true,
    default: ({schedule, onBack, loading}: any) => (
        <div data-testid='schedule-details'>
            {loading && <div>{'Loading details...'}</div>}
            {schedule && (
                <>
                    <h2>{schedule.name}</h2>
                    <button onClick={onBack}>{'Back'}</button>
                </>
            )}
        </div>
    ),
}));

describe('PagerDutySidebar', () => {
    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('should render and load schedules on mount', async () => {
        const mockSchedules = {
            schedules: [
                {id: 'SCHED1', name: 'Primary On-Call'},
                {id: 'SCHED2', name: 'Secondary On-Call'},
            ],
        };

        mockClient.getSchedules.mockResolvedValueOnce(mockSchedules);

        render(<PagerDutySidebar theme={mockTheme}/>);

        // Should show loading state initially
        expect(screen.getByText('Loading schedules...')).toBeInTheDocument();

        // Wait for schedules to load
        await waitFor(() => {
            expect(screen.getByTestId('schedule-SCHED1')).toBeInTheDocument();
            expect(screen.getByTestId('schedule-SCHED2')).toBeInTheDocument();
        });

        expect(mockClient.getSchedules).toHaveBeenCalledTimes(1);
    });

    it('should handle error when loading schedules fails', async () => {
        mockClient.getSchedules.mockRejectedValueOnce(new Error('API Error'));

        render(<PagerDutySidebar theme={mockTheme}/>);

        await waitFor(() => {
            expect(screen.getByText('Error: API Error')).toBeInTheDocument();
        });
    });

    it('should show schedule details when a schedule is clicked', async () => {
        const mockSchedules = {
            schedules: [
                {id: 'SCHED1', name: 'Primary On-Call'},
            ],
        };

        const mockScheduleDetails = {
            id: 'SCHED1',
            name: 'Primary On-Call',
            time_zone: 'America/New_York',
        };

        mockClient.getSchedules.mockResolvedValueOnce(mockSchedules);
        mockClient.getScheduleDetails.mockResolvedValueOnce(mockScheduleDetails);

        render(<PagerDutySidebar theme={mockTheme}/>);

        // Wait for schedules to load
        await waitFor(() => {
            expect(screen.getByTestId('schedule-SCHED1')).toBeInTheDocument();
        });

        // Click on a schedule
        fireEvent.click(screen.getByTestId('schedule-SCHED1'));

        // Should show loading state
        expect(screen.getByText('Loading details...')).toBeInTheDocument();

        // Wait for details to load
        await waitFor(() => {
            expect(screen.getByTestId('schedule-details')).toBeInTheDocument();

            // Check that the schedule name appears in the header
            const header = screen.getByRole('heading', {level: 3});
            expect(header).toHaveTextContent('Primary On-Call');
        });

        expect(mockClient.getScheduleDetails).toHaveBeenCalledWith('SCHED1');
    });

    it('should go back to list view when back is clicked', async () => {
        const mockSchedules = {
            schedules: [
                {id: 'SCHED1', name: 'Primary On-Call'},
            ],
        };

        const mockScheduleDetails = {
            id: 'SCHED1',
            name: 'Primary On-Call',
        };

        mockClient.getSchedules.mockResolvedValueOnce(mockSchedules);
        mockClient.getScheduleDetails.mockResolvedValueOnce(mockScheduleDetails);

        render(<PagerDutySidebar theme={mockTheme}/>);

        // Wait for schedules to load and click one
        await waitFor(() => {
            expect(screen.getByTestId('schedule-SCHED1')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByTestId('schedule-SCHED1'));

        // Wait for details to load
        await waitFor(() => {
            expect(screen.getByTestId('schedule-details')).toBeInTheDocument();
        });

        // Click back button
        fireEvent.click(screen.getByText('Back'));

        // Should show schedule list again
        expect(screen.getByTestId('schedule-list')).toBeInTheDocument();
        expect(screen.queryByTestId('schedule-details')).not.toBeInTheDocument();
    });
});
