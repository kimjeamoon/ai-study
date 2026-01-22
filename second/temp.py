```python
def fibonacci_loop(n):
    """
    반복문을 사용하여 피보나치 수열의 처음 n개 항을 생성하는 함수.

    Args:
        n (int): 생성할 피보나치 수열의 항의 개수.

    Returns:
        list: 피보나치 수열의 처음 n개 항을 담은 리스트.
              n이 0 이하일 경우 빈 리스트를 반환합니다.
    """
    if n <= 0:
        return []
    elif n == 1:
        return [0]
    else:
        fib_sequence = [0, 1]
        # 이미 2개의 항이 있으므로, n-2번 반복하여 나머지 항을 계산합니다.
        for _ in range(2, n):
            next_fib = fib_sequence[-1] + fib_sequence[-2]
            fib_sequence.append(next_fib)
        return fib_sequence
```