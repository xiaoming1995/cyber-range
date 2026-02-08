// 真实 API 的题目列表查询 hook
import { useEffect, useState } from 'react';
import * as adminAPI from '../api/admin';


export type Challenge = adminAPI.Challenge;

export interface UseChallengesResult {
    challenges: Challenge[];
    total: number;
    loading: boolean;
    error: string | null;
    refetch: () => void;
}

export function useChallenges(query: adminAPI.ChallengeListQuery = {}): UseChallengesResult {
    const [challenges, setChallenges] = useState<Challenge[]>([]);
    const [total, setTotal] = useState(0);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [refreshKey, setRefreshKey] = useState(0);

    useEffect(() => {
        setLoading(true);
        setError(null);

        adminAPI.listChallenges(query)
            .then((response) => {
                setChallenges(response.list);
                setTotal(response.total);
            })
            .catch((err) => {
                setError(err.message || '加载失败');
                setChallenges([]);
                setTotal(0);
            })
            .finally(() => {
                setLoading(false);
            });
    }, [query.page, query.pageSize, query.category, query.difficulty, query.status, query.search, refreshKey]);

    const refetch = () => setRefreshKey((k) => k + 1);

    return { challenges, total, loading, error, refetch };
}
