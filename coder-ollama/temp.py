import random
import math

def is_prime(n):
    if n <= 1:
        return False
    for i in range(2, int(math.sqrt(n)) + 1):
        if n % i == 0:
            return False
    return True

num = random.randint(1, 100)
print(f"Generated number: {num}")
if is_prime(num):
    print(f"{num} is a prime number")
else:
    print(f"{num} is not a prime number")