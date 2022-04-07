import subprocess
import time

durations=1 #minute
interval=10 #seconds
usageMem=[]
usageCPU=[]

def mem():
    mem = subprocess.check_output("kubectl top pods | awk '/nginx/{print $3}' | sed 's/Mi//g'", shell=True)
    usageMem.append(int(mem.decode()))
    
def cpu():
    cpu = subprocess.check_output("kubectl top pods | awk '/nginx/{print $2}' | sed 's/m//g'", shell=True)
    usageCPU.append(int(cpu.decode()))

def average():
    mem=0
    cpu=0
    for i in usageMem:
        mem+=i
    for j in usageCPU:
        cpu+=j
    return(mem/len(usageMem), cpu/len(usageCPU))

def highest():
    highest=0
    for i in usageMem:
        if i > highest:
            highest=i
    return highest

def anomaly(average, highest):
    if highest > (2*average):
        return True
    else:
        return False

while len(usageMem)<(durations*60/interval):
    mem()
    cpu()
    time.sleep(interval)

print(usageMem)
print(usageCPU)
print("CPU Average:", average()[1])
print("Memory Average:", average()[0])
print("Highest Memory:", highest())
print(anomaly(average()[0], highest()))