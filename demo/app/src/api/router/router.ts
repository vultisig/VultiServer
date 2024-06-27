import { endPoints } from "../endpoints";


export function route(id: string): Promise<Response> {
    return fetch(`${endPoints.router}/${id}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    });
}

export function routeStart(id: string, str: string[]): Promise<Response> {
    return fetch(`${endPoints.routerStart}/${id}`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(str)
    });
}

export function routeComplete(id: string): Promise<Response> {
    return fetch(`${endPoints.routerComplete}/${id}`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    });
}