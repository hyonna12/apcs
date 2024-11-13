from pymodbus.server import StartTcpServer
from pymodbus.datastore import ModbusSequentialDataBlock, ModbusSlaveContext, ModbusServerContext
import threading
import time

# 글로벌 컨텍스트 변수
global_context = None
server_running = True

def run_server():
    global global_context
    # 시작 주소를 0으로 하고, D5999까지 커버할 수 있도록 범위 설정
    store = ModbusSlaveContext(
        hr=ModbusSequentialDataBlock(0, [0]*6000)  # 0~5999 범위 할당
    )
    global_context = ModbusServerContext(slaves=store, single=True)

    # 모니터링 스레드 시작
    monitor_thread = threading.Thread(target=monitor_registers)
    monitor_thread.daemon = True
    monitor_thread.start()

    print("Starting Modbus Server on localhost:502")
    print("Memory range: D0000-D5999")
    server_thread = threading.Thread(target=StartTcpServer, 
                                   kwargs={'context': global_context, 'address': ("localhost", 502)})
    server_thread.daemon = True
    server_thread.start()

def monitor_registers():
    """레지스터 모니터링 및 자동 리셋"""
    prev_values = {}  # 이전 값 저장용
    
    while server_running:
        try:
            # D5500 영역 모니터링 (명령 레지스터)
            cmd_values = global_context[0].getValues(3, 5500, 10)
            for i, value in enumerate(cmd_values):
                if value != 0:
                    cmd_addr = 5500 + i
                    print(f"\n=== Command Received ===")
                    print(f"D{cmd_addr} = {bin(value)[2:]:016} (Command)")
                    
                    # 명령 수신 시 해당하는 D5000 영역 비트 설정
                    status_addr = 5000 + (cmd_addr - 5500)  # 정확한 상태 주소 계산
                    global_context[0].setValues(3, status_addr, [value])
                    print(f"D{status_addr} = {bin(value)[2:]:016} (Status)")
                    
                    # 명령 레지스터 리셋
                    global_context[0].setValues(3, cmd_addr, [0])
                    print(f"D{cmd_addr} reset to 0")
                    print("=== Command Complete ===\n")

            # D5000 영역 모니터링 (상태 레지스터)
            status_values = global_context[0].getValues(3, 5000, 10)
            for i, value in enumerate(status_values):
                status_addr = 5000 + i
                # 이전 값과 다른 경우에만 출력
                if status_addr not in prev_values or prev_values[status_addr] != value:
                    if value != 0:  # 0이 아닌 값만 출력
                        print(f"Status Change: D{status_addr} = {bin(value)[2:]:016}")
                    prev_values[status_addr] = value

        except Exception as e:
            pass
        time.sleep(0.1)

def print_menu():
    print("\n=== Modbus Server Menu ===")
    print("1. Read register value")
    print("2. Write register value")
    print("3. Monitor register")
    print("4. Exit")
    return input("Select option: ")

def read_register(address):
    try:
        # 16진수 문자열에서 'D' 제거하고 변환
        if isinstance(address, str):
            address = address.replace('D', '')
            address = int(address, 16)  # 16진수 문자열을 정수로 변환
            
        print(f"\nReading D{address:04X}")
        value = global_context[0].getValues(3, address, 1)[0]
        print(f"D{address:04X} = {bin(value)} (0x{value:04X})")
        return value
    except Exception as e:
        print(f"Error reading D{address:04X}: {e}")

def write_register(address, value):
    try:
        global_context[0].setValues(3, address, [value])
        print(f"Written {bin(value)} to register {address:04X}h")
    except Exception as e:
        print(f"Error writing register: {e}")

if __name__ == "__main__":
    run_server()
    
    while True:
        try:
            option = print_menu()
            if option == "1":
                addr = int(input("Enter register address : "))
                read_register(addr)
            elif option == "2":
                addr = input("Enter register address: ")
                val = int(input("Enter value (decimal): "))
                write_register(addr, val)
            elif option == "3":
                addr = input("Enter register address: ")
                print(f"Monitoring register D{addr} (Ctrl+C to stop)")
                while True:
                    read_register(addr)
                    time.sleep(1)
            elif option == "4":
                print("Exiting...")
                server_running = False
                break
        except KeyboardInterrupt:
            continue
        except ValueError as e:
            print(f"Invalid input: {e}")
        except Exception as e:
            print(f"Error: {e}") 