// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import manifest from '@/manifest';

export class Client {
    private baseUrl: string;

    constructor() {
        // Use window.location.origin to construct the base URL
        const siteUrl = window.location.origin;
        this.baseUrl = `${siteUrl}/plugins/${manifest.id}/api/v1`;
    }

    async getSchedules() {
        const response = await fetch(`${this.baseUrl}/schedules`, {
            method: 'GET',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to fetch schedules');
        }

        return response.json();
    }

    async getOnCalls(scheduleId?: string) {
        const params = scheduleId ? `?schedule_id=${scheduleId}` : '';
        const response = await fetch(`${this.baseUrl}/oncalls${params}`, {
            method: 'GET',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to fetch on-calls');
        }

        return response.json();
    }

    async getScheduleDetails(scheduleId: string) {
        const response = await fetch(`${this.baseUrl}/schedule?id=${scheduleId}`, {
            method: 'GET',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to fetch schedule details');
        }

        return response.json();
    }

    async getServices() {
        const response = await fetch(`${this.baseUrl}/services`, {
            method: 'GET',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to fetch services');
        }

        return response.json();
    }

    async createIncident(title: string, description: string, serviceId: string, assigneeIds?: string[]) {
        const body = {
            title,
            description,
            service_id: serviceId,
            assignee_ids: assigneeIds || [],
        };

        const response = await fetch(`${this.baseUrl}/incidents`, {
            method: 'POST',
            credentials: 'include',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(body),
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.message || 'Failed to create incident');
        }

        return response.json();
    }
}

const client = new Client();
export default client;
