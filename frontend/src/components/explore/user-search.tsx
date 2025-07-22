import { useState } from "react";
import { Input } from "../ui/input";
import { UserAvatar } from "../ui/user-avatar";
import { SkeletonList } from "../ui/skeleton-list";
import { userAPI } from "@/lib/api/service/user";
import { Search, X } from "lucide-react";
import { useDebounce } from "@/hooks/use-debounce";
import { useQuery } from "@tanstack/react-query";
import type { UserSearchItem } from "@/lib/api/types";
import { Button } from "../ui/button";

const MIN_QUERY_LENGTH = 2;

interface UserSearchProps {
  onUserSelect: (user: UserSearchItem) => void;
}

export function UserSearch({ onUserSelect }: UserSearchProps) {
  const [query, setQuery] = useState("");
  const [isFocused, setIsFocused] = useState(false);
  const debouncedQuery = useDebounce(query, 300);

  const { data: results, isLoading } = useQuery({
    queryKey: ["userSearch", debouncedQuery],
    queryFn: async () => {
      const response = await userAPI.searchUsers(debouncedQuery);
      if (!response.success || !response.data) {
        throw new Error(response.message || "Search failed");
      }
      return response.data;
    },
    enabled: debouncedQuery.trim().length >= MIN_QUERY_LENGTH,
    staleTime: 1000 * 60 * 5,
  });

  const handleUserSelect = (user: UserSearchItem) => {
    onUserSelect(user);
    setQuery("");
    setIsFocused(false);
  };

  return (
    <div className="relative w-full">
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
        <Input
          type="search"
          placeholder="Search for players..."
          className="w-full pl-10 pr-4 py-2 text-base"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={() => setIsFocused(true)}
          onBlur={() => setTimeout(() => setIsFocused(false), 150)}
        />
        {query.length > 0 && (
          <Button
            variant="ghost"
            size="icon"
            className="absolute right-2 top-1/2 -translate-y-1/2 h-7 w-7 rounded-full cursor-pointer"
            onClick={() => setQuery("")}
          >
            <X className="h-4 w-4" />
            <span className="sr-only">Clear search</span>
          </Button>
        )}
      </div>

      {isFocused &&
        (query.length > 0 || isLoading || (results && results?.length)) && (
          <div className="absolute z-50 mt-1 w-full bg-card border rounded-md shadow-lg max-h-60 overflow-y-auto">
            {isLoading && query.length > 0 && <SkeletonList count={3} />}
            {!isLoading &&
              debouncedQuery.length > 0 &&
              results?.length === 0 && (
                <div className="p-4 text-center text-sm text-muted-foreground">
                  No users found matching "{debouncedQuery}"
                </div>
              )}
            {!isLoading && results && results.length > 0 && (
              <ul>
                {results.map((user) => (
                  <li
                    key={user.user_id}
                    className="px-3 py-2 hover:bg-muted cursor-pointer rounded-md"
                    onMouseDown={() => handleUserSelect(user)}
                  >
                    <div className="flex items-center space-x-3">
                      <UserAvatar
                        userId={user.user_id}
                        displayName={user.avatar_url || user.username}
                        avatarUrl={user.avatar_url}
                        updatedAt={user.updated_at}
                      />
                      <div>
                        <p className="text-sm font-medium leading-none">
                          {user.display_name || user.username}
                        </p>
                        {user.display_name && (
                          <p className="text-xs text-muted-foreground">
                            @{user.username}
                          </p>
                        )}
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}
    </div>
  );
}
