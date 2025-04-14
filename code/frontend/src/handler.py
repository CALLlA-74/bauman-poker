from schemas.ws_from_player import AuthMessage, ActionMessage, PongMessage
from config import BASE_URL, TOKENS

from time import time

import requests

from color import FIELD_COLOR2

from tkinter import messagebox
from tkinter import *

players = []
info = []
cards_dict = {}

poker_table = Tk()

tcard1 = Label(poker_table, bg=FIELD_COLOR2)
tcard2 = Label(poker_table, bg=FIELD_COLOR2)
tcard3 = Label(poker_table, bg=FIELD_COLOR2)
tcard4 = Label(poker_table, bg=FIELD_COLOR2)
tcard5 = Label(poker_table, bg=FIELD_COLOR2)

pcard1 = Label(poker_table, bg=FIELD_COLOR2)
pcard2 = Label(poker_table, bg=FIELD_COLOR2)

call_img = PhotoImage(file="img/buttons/call.png")
a_call_img = PhotoImage(file="img/buttons/call_a.png")
check_img = PhotoImage(file="img/buttons/check.png")
a_check_img = PhotoImage(file="img/buttons/check_a.png")
fold_img = PhotoImage(file="img/buttons/fold.png")
a_fold_img = PhotoImage(file="img/buttons/fold_a.png")
raise1_img = PhotoImage(file="img/buttons/raise1.png")
a_raise1_img = PhotoImage(file="img/buttons/raise1_a.png")
raise2_img = PhotoImage(file="img/buttons/raise2.png")
a_raise2_img = PhotoImage(file="img/buttons/raise2_a.png")
allin_img = PhotoImage(file="img/buttons/allin.png")
a_allin_img = PhotoImage(file="img/buttons/allin_a.png")

cover_img = PhotoImage(file="img/pack/cover.png")

def on_enter_call(e):
    call_btn['image'] = a_call_img
def on_leave_call(e):
    call_btn['image'] = call_img

def on_enter_check(e):
    check_btn['image'] = a_check_img
def on_leave_check(e):
    check_btn['image'] = check_img

def on_enter_fold(e):
    fold_btn['image'] = a_fold_img
def on_leave_fold(e):
    fold_btn['image'] = fold_img

def on_enter_raise1(e):
    raise1_btn['image'] = a_raise1_img
def on_leave_raise1(e):
    raise1_btn['image'] = raise1_img

def on_enter_raise2(e):
    raise2_btn['image'] = a_raise2_img
def on_leave_raise2(e):
    raise2_btn['image'] = raise2_img  

def on_enter_allin(e):
    allin_btn['image'] = a_allin_img
def on_leave_allin(e):
    allin_btn['image'] = allin_img

def on_disable(e):
    pass

def disable(btns):
    for btn in btns:
        btn['state'] = DISABLED
        btn.bind("<Enter>", on_disable)

call_btn = Button(poker_table, 
                image=call_img, 
                bg=FIELD_COLOR2, 
                activebackground=FIELD_COLOR2,
                relief = FLAT, bd=0,
                command=lambda:disable([call_btn, check_btn, fold_btn]))
call_btn.bind("<Enter>", on_enter_call)
call_btn.bind("<Leave>", on_leave_call)
call_btn.place(x=850, y= 520)

check_btn = Button(poker_table, 
                image=check_img, 
                bg=FIELD_COLOR2, 
                activebackground=FIELD_COLOR2,
                relief = FLAT, bd=0)
check_btn.bind("<Enter>", on_enter_check)
check_btn.bind("<Leave>", on_leave_check)
check_btn.place(x=850, y= 590)

fold_btn = Button(poker_table, 
                image=fold_img, 
                bg=FIELD_COLOR2, 
                activebackground=FIELD_COLOR2,
                relief = FLAT, bd=0)
fold_btn.bind("<Enter>", on_enter_fold)
fold_btn.bind("<Leave>", on_leave_fold)
fold_btn.place(x=850, y= 660)

raise1_btn = Button(poker_table, 
                image=raise1_img, 
                bg=FIELD_COLOR2, 
                activebackground=FIELD_COLOR2,
                relief = FLAT, bd=0)
raise1_btn.bind("<Enter>", on_enter_raise1)
raise1_btn.bind("<Leave>", on_leave_raise1)
raise1_btn.place(x=1020, y= 520)

raise2_btn = Button(poker_table, 
                image=raise2_img, 
                bg=FIELD_COLOR2, 
                activebackground=FIELD_COLOR2,
                relief = FLAT, bd=0)
raise2_btn.bind("<Enter>", on_enter_raise2)
raise2_btn.bind("<Leave>", on_leave_raise2)
raise2_btn.place(x=1020, y= 590)

allin_btn = Button(poker_table, 
                image=allin_img, 
                bg=FIELD_COLOR2, 
                activebackground=FIELD_COLOR2,
                relief = FLAT, bd=0)
allin_btn.bind("<Enter>", on_enter_allin)
allin_btn.bind("<Leave>", on_leave_allin)
allin_btn.place(x=1020, y= 660)

pong_msg = PongMessage().model_dump_json()

def show_players():
    pass

def show_info():
    pass

def get_player_info(user_uid):
    params = {'userUid': user_uid}
    response = requests.get(BASE_URL+'/poker/v1/players', 
                            params=params,
                            headers={'Authorization':TOKENS[0]})
    
    if response.status_code == 200:
        player_info = response.json()
        username = player_info['Username']
        rank = player_info['UserRank']
        
        
        player = [user_uid, username, rank, '10000', '0', False, False]
        players.append(player)
        show_players()



    


message = ''

if message['MessageType'] == 'PING':
    pass # SEND PONG
elif message['MessageType'] == 'ACK':
    pass # ERRORS HANDLER
elif message['MessageType'] == 'EVENT':
    
    if message['EventType'] == 'PLAYER-ACTION-EVENT':
        
        user_uid = message['EventDescriptor']['UserUid']
        
        if message['EventDescriptor']['ActionType'] == 'INCOME':
            get_player_info(user_uid)
        
        elif message['EventDescriptor']['ActionType'] == 'OUTCOME':
            for player in players:
                if user_uid == player[0]:
                    player[6] = True
            if user_uid == TOKENS[2]:
                messagebox.showerror('YOU LOOSE', 'К сожалению, вы проиграли!')
                poker_table.destroy()
        
        elif message['EventDescriptor']['ActionType'] == 'BOUT':
            if user_uid == TOKENS[2]:
                info[3] = message['EventDescriptor']['BestCombName']
                for bout_variant in message['EventDescriptor']['BoutVariants']:

                    # РАЗБЛОКИРОВАТЬ КНОПКИ
                    if bout_variant['VariantType'] == 'FOLD':
                        fold_btn.config(state=NORMAL)
                        fold_btn.bind("<Enter>", on_enter_fold)
                        fold_btn.bind("<Leave>", on_leave_fold)
                    elif bout_variant['VariantType'] == 'CHECK':
                        check_btn.config(state=NORMAL)
                        check_btn.bind("<Enter>", on_enter_check)
                        check_btn.bind("<Leave>", on_leave_check)
                    elif bout_variant['VariantType'] == 'CALL':
                        call_btn.config(state=NORMAL)
                        call_btn.bind("<Enter>", on_enter_call)
                        call_btn.bind("<Leave>", on_leave_call)
                    elif bout_variant['VariantType'] == 'RAISE':
                        
                        if bout_variant['RaiseVAriants'] == 'X1_5':
                            raise1_btn.config(state=NORMAL)
                            raise1_btn.bind("<Enter>", on_enter_raise1)
                            raise1_btn.bind("<Leave>", on_leave_raise1)
                        elif bout_variant['RaiseVAriants'] == 'X2':
                            raise2_btn.config(state=NORMAL)
                            raise2_btn.bind("<Enter>", on_enter_raise2)
                            raise2_btn.bind("<Leave>", on_leave_raise2)
                        elif bout_variant['RaiseVAriants'] == 'ALL-IN':
                            allin_btn.config(state=NORMAL)
                            allin_btn.bind("<Enter>", on_enter_allin)
                            allin_btn.bind("<Leave>", on_leave_allin)
                show_info()
        
        elif message['EventDescriptor']['ActionType'] == 'FOLD':
            for player in players:
                if user_uid == player[0]:
                    player[6] = False
            show_players()
        
        elif message['EventDescriptor']['ActionType'] == 'CHECK':
            pass
        
        elif message['EventDescriptor']['ActionType'] == 'CALL':
            for player in players:
                if user_uid == player[0]:
                    player[3] = str(message['EventDescriptor']['NewDeposit'])
                    player[4] = str(message['EventDescriptor']['NewBet'])
            if user_uid == TOKENS[2]:
                info[1] = str(message['EventDescriptor']['NewDeposit'])
                info[2] = str(message['EventDescriptor']['NewBet'])
            show_info()
        
        elif message['EventDescriptor']['ActionType'] == 'RAISE':
            for player in players:
                if user_uid == player[0]:
                    player[3] = str(message['EventDescriptor']['NewDeposit'])
                    player[4] = str(message['EventDescriptor']['NewBet'])
            if user_uid == TOKENS[2]:
                info[1] = str(message['EventDescriptor']['NewDeposit'])
                info[2] = str(message['EventDescriptor']['NewBet'])
            show_info()
        
        elif message['EventDescriptor']['ActionType'] == 'ALL-IN':
            for player in players:
                if user_uid == player[0]:
                    player[3] = str(message['EventDescriptor']['NewDeposit'])
                    player[4] = str(message['EventDescriptor']['NewBet'])
            if user_uid == TOKENS[2]:
                info[1] = str(message['EventDescriptor']['NewDeposit'])
                info[2] = str(message['EventDescriptor']['NewBet'])
            show_info()
        
        elif message['EventDescriptor']['ActionType'] == 'SET-DEALER':
            pass
        
        elif message['EventDescriptor']['ActionType'] == 'MIN-BLIND-IN':
            for player in players:
                if user_uid == player[0]:
                    player[3] = str(message['EventDescriptor']['NewDeposit'])
                    player[4] = str(message['EventDescriptor']['NewBet'])
            if user_uid == TOKENS[2]:
                info[1] = str(message['EventDescriptor']['NewDeposit'])
                info[2] = str(message['EventDescriptor']['NewBet'])
            show_info()
        
        elif message['EventDescriptor']['ActionType'] == 'MAX-BLIND-IN':
            for player in players:
                if user_uid == player[0]:
                    player[3] = str(message['EventDescriptor']['NewDeposit'])
                    player[4] = str(message['EventDescriptor']['NewBet'])
            if user_uid == TOKENS[2]:
                info[1] = str(message['EventDescriptor']['NewDeposit'])
                info[2] = str(message['EventDescriptor']['NewBet'])
            show_info()
    
    
    
    elif message['EventType'] == 'GAME-EVENT':
        
        if message['EventDescriptor']['ActionType'] == 'ROOM_STATE_UPDATE':
            if message['EventDescriptor']['NewRoomState'] == 'GAMING':
                tcard1.config(image=cover_img)
                tcard2.config(image=cover_img)
                tcard3.config(image=cover_img)
                tcard4.config(image=cover_img)
                tcard5.config(image=cover_img)
                pcard1.config(image=cover_img)
                pcard2.config(image=cover_img)
            elif message['EventDescriptor']['NewRoomState'] == 'DISSLOLUTION':
                messagebox.showerror('YOU WIN', 'Поздравляем, вы победили!')
        
        elif message['EventDescriptor']['ActionType'] == 'NEW_ROUND':
            tcard1.config(image=cover_img)
            tcard2.config(image=cover_img)
            tcard3.config(image=cover_img)
            tcard4.config(image=cover_img)
            tcard5.config(image=cover_img)
            pcard1.config(image=cover_img)
            pcard2.config(image=cover_img)
        
        elif message['EventDescriptor']['ActionType'] == 'NEW_TRADE_ROUND':
            pass
        
        elif message['EventDescriptor']['ActionType'] == 'PERSONAL_CARDS':
            suit1 = message['EventDescriptor']['PlayingCardsList'][0]['CardSuit']
            index1 = message['EventDescriptor']['PlayingCardsList'][0]['Index']
            suit2 = message['EventDescriptor']['PlayingCardsList'][0]['CardSuit']
            index2 = message['EventDescriptor']['PlayingCardsList'][0]['Index']

            pcard1.config(image=cards_dict[suit1][index1])
            pcard2.config(image=cards_dict[suit2][index2])
        
        
        elif message['EventDescriptor']['ActionType'] == 'CARDS_ON_TABLE':
            i = 1
            for card in message['EventDescriptor']['PlayingCardsList']:
                suit = card['CardSuit']
                index = card['Index']
                if i == 1:
                    tcard1.config(image=cards_dict[suit][index])
                elif i == 2:
                    tcard2.config(image=cards_dict[suit][index])
                elif i == 3:
                    tcard3.config(image=cards_dict[suit][index])
                elif i == 4:
                    tcard4.config(image=cards_dict[suit][index])
                elif i == 5:
                    tcard5.config(image=cards_dict[suit][index])
                i+=1
        
        
        elif message['EventDescriptor']['ActionType'] == 'BET_ACCEPTED':
            for player in players:
                player[4] = ''
            info[0] = str(message['EventDescriptor']['NewStack'])
            show_players()
            show_info()
        
        
        elif message['EventDescriptor']['ActionType'] == 'WINNER_RESULT':
            winners = ''
            for uuid in message['EventDescriptor']['WinnerUids']:
                for player in players:
                    if uuid == player[0]:
                        winners+= ' ' + player[1]
            combo = message['EventDescriptor']['BestCombName']
            info[0] = '0'          
            
            show_players()
            show_info()
            messagebox.showwarning('WINNER(S)', 'Поебедитель(-и):' + winners 
                                   + '\nКомбинация: ' + combo)











room_uid = 'e0358117-9704-49b8-8d4e-b62133114f1e'
user_uid = 'e0358117-9704-49b8-8d4e-b62133114f1e'
last_event_id = 228

pong_msg = PongMessage().model_dump_json()
print(pong_msg)


auth_msg = AuthMessage(MessageId=int(time()*1000),
                       RoomUid=room_uid,
                       Token=TOKENS[0],
                       LastEventId=last_event_id
                       ).model_dump_json()
print(auth_msg)

act_msg1 = ActionMessage(MessageType='VOTE',
                        MessageId=int(time()*1000),
                        RoomUid=room_uid,
                        UserUid=user_uid,
                        VoteType='WAIT'
                        ).model_dump_json()
print(act_msg1)

act_msg2 = ActionMessage(MessageType='GAME-ACTION',
                        MessageId=int(time()*1000),
                        RoomUid=room_uid,
                        UserUid=user_uid,
                        ActionType='CHECK'
                        ).model_dump_json()
print(act_msg2)

act_msg3 = ActionMessage(MessageType='GAME-ACTION',
                        MessageId=int(time()*1000),
                        RoomUid=room_uid,
                        UserUid=user_uid,
                        ActionType='RAISE',
                        Coef='ALL-IN'
                        ).model_dump_json()
print(act_msg3)