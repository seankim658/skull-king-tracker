import { useState, useCallback, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Input } from "../ui/input";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import { Skeleton } from "../ui/skeleton";
import { userAPI } from "@/lib/api/service/user";
import type { UserSearchItem } from "@/lib/api/types";
import { errorExtract, getFullAvatarURL, getAvatarFallback } from "@/lib/utils";
import { Search } from "lucide-react";
import { useDebounce } from "@/hooks/use-debounce";

const MIN_QUERY_LENGTH = 2;

export function UserSearch() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<UserSearchItem[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isFocused, setIsFocused] = useState(false);
  const debouncedQuery = useDebounce(query, 300);
  const navigate = useNavigate();

  const fetchUsers = useCallback(async (searchQuery: string) => {
    if (searchQuery.trim().length < MIN_QUERY_LENGTH) {
      setResults([]);
      setIsLoading(false);
      return;
    }
    setIsLoading(true);
    try {
      const response = await userAPI.searchUsers(searchQuery, 5);
      if (response.success && response.data) {
        setResults(response.data);
      } else {
        console.error(response.message || "Failed to search users");
        setResults([]);
      }
    } catch (e) {
      console.error(errorExtract(e, "Could not fetch users"));
      setResults([]);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (debouncedQuery.trim().length >= MIN_QUERY_LENGTH) {
      fetchUsers(debouncedQuery);
    } else {
      setResults([]);
      setIsLoading(false);
    }
  }, [debouncedQuery, fetchUsers]);

  const handleUserSelect = (userId: string) => {
    setQuery("");
    setResults([]);
    setIsFocused(false);
    navigate(`/users/${userId}`);
  };

  return (
    <div className="relative w-full max-w-xl mx-auto">
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
      </div>

      {isFocused && (query.length > 0 || isLoading || results.length > 0) && (
        <div className="absolute z-10 mt-1 w-full bg-card border rounded-md shadow-lg max-h-80 overflow-y-auto">
          {isLoading && query.length > 0 && (
            <div className="p-2">
              {[...Array(3)].map((_, i) => (
                <div key={i} className="flex items-center space-x-3 p-2">
                  <Skeleton className="h-10 w-10 rounded-full" />
                  <div className="space-y-1">
                    <Skeleton className="h-4 w-32" />
                    <Skeleton className="h-3 w-24" />
                  </div>
                </div>
              ))}
            </div>
          )}
          {!isLoading && debouncedQuery.length > 0 && results.length === 0 && (
            <div className="p-4 text-center text-sm text-muted-foreground">
              No users found matching "{debouncedQuery}"
            </div>
          )}
          {!isLoading && results.length > 0 && (
            <ul>
              {results.map((user) => (
                <li
                  key={user.user_id}
                  className="px-3 py-2 hover:bg-muted cursor-pointer rounded-md"
                  onMouseDown={() => handleUserSelect(user.user_id)}
                >
                  <div className="flex items-center space-x-3">
                    <Avatar className="h-10 w-10">
                      <AvatarImage
                        src={getFullAvatarURL(user.avatar_url)}
                        alt={user.username}
                      />
                      <AvatarFallback>
                        {getAvatarFallback(user.display_name || user.username)}
                      </AvatarFallback>
                    </Avatar>
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
