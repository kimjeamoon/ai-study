import random
import math

def is_prime(num):
    if num < 2:
        return False
    for i in range(2, int(math.sqrt(num)) + 1):
        if num % i == 0:
            return False
    return True

random_number = random.randint(1, 100)
if is_prime(random_number):
    print(f"The generated number {random_number} is a prime number.")
else:
    print(f"The generated number {random_number} is not a prime number.")