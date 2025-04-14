from tkinter import *
from tkinter import messagebox
from tkinter import font
from PIL import ImageTk, Image

from time import time, sleep

from color import FIELD_COLOR, FIELD_COLOR2, FONT_COLOR, PLAYER_COL_COLOR
from config import BASE_URL, TOKENS, WS_BASE_URL

from schemas.ws_from_player import AuthMessage, ActionMessage, PongMessage

pong_msg = PongMessage().model_dump_json()

from threading import Thread
import requests
import websocket
import json

# ИГРОКИ
# players = [
#     ['sdkfs-dvwej-bwevw-eb', 'Player1', 'ПОЛКОВНИК', '500', '1250'],
#     ['Player2', 'РЯДОВОЙ', '5000', '2500'],
#     ['Player3', 'ГЕНЕРАЛ', '12500', '0'],
#     ['Player4', 'КАПИТАН', '7250', '0'],
# ]
player = ['', '', '', '', '', False, False]
players = []

# ОСНОВНАЯ ИНФОРМАЦИЯ
info = [
    '0',
    '10 000',
    '0',
    ''
]

auth = [True]
vote = [True, False]
action = ['']
delay = [False]

def poker_table():
    
    room_info_response = requests.get(BASE_URL+'/poker/v1/rooms/matching', 
                                      headers={'Authorization':TOKENS[0]})
    
    if room_info_response.status_code == 401:
        messagebox.showerror('Error 401', 'Время действия AccessToken истекло!')
        return
    elif room_info_response.status_code == 500:
        messagebox.showerror('Error 500', 'Внутренняя ошибка сервера!')
        return 
    
    room_info = room_info_response.json()
    room_uid = room_info['RoomUid']
    last_event_id = room_info['LastEventId']

    for room_player in room_info['PlayerList']:
        player[0] = room_player['UserUid']
        player[1] = room_player['Username']
        player[2] = room_player['UserRank']
        player[3] = str(room_player['Deposit'])
        player[4] = str(room_player['Bet'])
        print(player)
        player_c = player.copy()
        players.append(player_c)


    poker_table = Toplevel()
    # poker_table = Tk()
    poker_table.grab_set()
    poker_table.lift()

    # ШРИФТЫ
    font_player = font.Font(family="Century Gothic", size=20, weight='bold')
    font_bet = font.Font(family="Century Gothic", size=20)
    font_bet_big = font.Font(family="Century Gothic", size=30, weight='bold')
    
    width = poker_table.winfo_screenwidth() - 10
    height = poker_table.winfo_screenheight() - 30
    
    poker_table.iconbitmap("icon.ico")
    poker_table.geometry('%dx%d+0+0' % (width, height))
    poker_table.title('TEXAS HOLDEM -- '+TOKENS[3])
    poker_table.resizable(True, True)
    poker_table.configure(background = FIELD_COLOR2)

    table_img = PhotoImage(file="img/table.png")
    table = Label(poker_table, image=table_img, bg=FIELD_COLOR2)
    table.place(x= 285, y= 20)

    bet_img = PhotoImage(file="img/money.png")
    bet_label = Label(poker_table, image=bet_img, bg=FIELD_COLOR2)
    bet_label.place(x= 666, y= 75)

    # РАНГИ
    recruit_img = PhotoImage(file="img/ranks/recruit.png")
    soldier_img = PhotoImage(file="img/ranks/soldier.png")
    sergeant_img = PhotoImage(file="img/ranks/sergeant.png")
    captain_img = PhotoImage(file="img/ranks/captain.png")
    major_img = PhotoImage(file="img/ranks/major.png")
    colonel_img = PhotoImage(file="img/ranks/colonel.png")
    general_img = PhotoImage(file="img/ranks/general.png")

    ranks_dict = {
        'РЕКРУТ': recruit_img,
        'РЯДОВОЙ': soldier_img,
        'СЕРЖАНТ': sergeant_img,
        'КАПИТАН': captain_img,
        'МАЙОР': major_img,
        'ПОЛКОВНИК': colonel_img,
        'ГЕНЕРАЛ': general_img
    }

    # КАРТЫ
    clubs_2_img = PhotoImage(file="img/pack/CLUBS/2.png")
    clubs_3_img = PhotoImage(file="img/pack/CLUBS/3.png")
    clubs_4_img = PhotoImage(file="img/pack/CLUBS/4.png")
    clubs_5_img = PhotoImage(file="img/pack/CLUBS/5.png")
    clubs_6_img = PhotoImage(file="img/pack/CLUBS/6.png")
    clubs_7_img = PhotoImage(file="img/pack/CLUBS/7.png")
    clubs_8_img = PhotoImage(file="img/pack/CLUBS/8.png")
    clubs_9_img = PhotoImage(file="img/pack/CLUBS/9.png")
    clubs_10_img = PhotoImage(file="img/pack/CLUBS/10.png")
    clubs_jack_img = PhotoImage(file="img/pack/CLUBS/JACK.png")
    clubs_queen_img = PhotoImage(file="img/pack/CLUBS/QUEEN.png")
    clubs_king_img = PhotoImage(file="img/pack/CLUBS/KING.png")
    clubs_ace_img = PhotoImage(file="img/pack/CLUBS/ACE.png")

    diamonds_2_img = PhotoImage(file="img/pack/DIAMONDS/2.png")
    diamonds_3_img = PhotoImage(file="img/pack/DIAMONDS/3.png")
    diamonds_4_img = PhotoImage(file="img/pack/DIAMONDS/4.png")
    diamonds_5_img = PhotoImage(file="img/pack/DIAMONDS/5.png")
    diamonds_6_img = PhotoImage(file="img/pack/DIAMONDS/6.png")
    diamonds_7_img = PhotoImage(file="img/pack/DIAMONDS/7.png")
    diamonds_8_img = PhotoImage(file="img/pack/DIAMONDS/8.png")
    diamonds_9_img = PhotoImage(file="img/pack/DIAMONDS/9.png")
    diamonds_10_img = PhotoImage(file="img/pack/DIAMONDS/10.png")
    diamonds_jack_img = PhotoImage(file="img/pack/DIAMONDS/JACK.png")
    diamonds_queen_img = PhotoImage(file="img/pack/DIAMONDS/QUEEN.png")
    diamonds_king_img = PhotoImage(file="img/pack/DIAMONDS/KING.png")
    diamonds_ace_img = PhotoImage(file="img/pack/DIAMONDS/ACE.png")

    hearts_2_img = PhotoImage(file="img/pack/HEARTS/2.png")
    hearts_3_img = PhotoImage(file="img/pack/HEARTS/3.png")
    hearts_4_img = PhotoImage(file="img/pack/HEARTS/4.png")
    hearts_5_img = PhotoImage(file="img/pack/HEARTS/5.png")
    hearts_6_img = PhotoImage(file="img/pack/HEARTS/6.png")
    hearts_7_img = PhotoImage(file="img/pack/HEARTS/7.png")
    hearts_8_img = PhotoImage(file="img/pack/HEARTS/8.png")
    hearts_9_img = PhotoImage(file="img/pack/HEARTS/9.png")
    hearts_10_img = PhotoImage(file="img/pack/HEARTS/10.png")
    hearts_jack_img = PhotoImage(file="img/pack/HEARTS/JACK.png")
    hearts_queen_img = PhotoImage(file="img/pack/HEARTS/QUEEN.png")
    hearts_king_img = PhotoImage(file="img/pack/HEARTS/KING.png")
    hearts_ace_img = PhotoImage(file="img/pack/HEARTS/ACE.png")

    spades_2_img = PhotoImage(file="img/pack/SPADES/2.png")
    spades_3_img = PhotoImage(file="img/pack/SPADES/3.png")
    spades_4_img = PhotoImage(file="img/pack/SPADES/4.png")
    spades_5_img = PhotoImage(file="img/pack/SPADES/5.png")
    spades_6_img = PhotoImage(file="img/pack/SPADES/6.png")
    spades_7_img = PhotoImage(file="img/pack/SPADES/7.png")
    spades_8_img = PhotoImage(file="img/pack/SPADES/8.png")
    spades_9_img = PhotoImage(file="img/pack/SPADES/9.png")
    spades_10_img = PhotoImage(file="img/pack/SPADES/10.png")
    spades_jack_img = PhotoImage(file="img/pack/SPADES/JACK.png")
    spades_queen_img = PhotoImage(file="img/pack/SPADES/QUEEN.png")
    spades_king_img = PhotoImage(file="img/pack/SPADES/KING.png")
    spades_ace_img = PhotoImage(file="img/pack/SPADES/ACE.png")

    cards_dict = {
        'CLUBS': {
            '2': clubs_2_img,
            '3': clubs_3_img,
            '4': clubs_4_img,
            '5': clubs_5_img,
            '6': clubs_6_img,
            '7': clubs_7_img,
            '8': clubs_8_img,
            '9': clubs_9_img,
            '10': clubs_10_img,
            'JACK': clubs_jack_img,
            'QUEEN': clubs_queen_img,
            'KING': clubs_king_img,
            'ACE': clubs_ace_img
        },
        'DIAMONDS': {
            '2': diamonds_2_img,
            '3': diamonds_3_img,
            '4': diamonds_4_img,
            '5': diamonds_5_img,
            '6': diamonds_6_img,
            '7': diamonds_7_img,
            '8': diamonds_8_img,
            '9': diamonds_9_img,
            '10': diamonds_10_img,
            'JACK': diamonds_jack_img,
            'QUEEN': diamonds_queen_img,
            'KING': diamonds_king_img,
            'ACE': diamonds_ace_img
        },
        'HEARTS': {
            '2': hearts_2_img,
            '3': hearts_3_img,
            '4': hearts_4_img,
            '5': hearts_5_img,
            '6': hearts_6_img,
            '7': hearts_7_img,
            '8': hearts_8_img,
            '9': hearts_9_img,
            '10': hearts_10_img,
            'JACK': hearts_jack_img,
            'QUEEN': hearts_queen_img,
            'KING': hearts_king_img,
            'ACE': hearts_ace_img
        },
        'SPADES': {
            '2': spades_2_img,
            '3': spades_3_img,
            '4': spades_4_img,
            '5': spades_5_img,
            '6': spades_6_img,
            '7': spades_7_img,
            '8': spades_8_img,
            '9': spades_9_img,
            '10': spades_10_img,
            'JACK': spades_jack_img,
            'QUEEN': spades_queen_img,
            'KING': spades_king_img,
            'ACE': spades_ace_img
        }
    }

    cover_img = PhotoImage(file="img/pack/cover.png")
    cover_img_mini = PhotoImage(file="img/pack/cover_mini.png")

    # КНОПКИ
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
    wait_img = PhotoImage(file="img/buttons/wait.png")
    begin_img = PhotoImage(file="img/buttons/begin.png")

    # ЛЕЙБЛЫ
        # КАРТЫ
    tcard1 = Label(poker_table, bg=FIELD_COLOR2)
    tcard2 = Label(poker_table, bg=FIELD_COLOR2)
    tcard3 = Label(poker_table, bg=FIELD_COLOR2)
    tcard4 = Label(poker_table, bg=FIELD_COLOR2)
    tcard5 = Label(poker_table, bg=FIELD_COLOR2)

    pcard1 = Label(poker_table, bg=FIELD_COLOR2)
    pcard2 = Label(poker_table, bg=FIELD_COLOR2)

    tcard1.place(x= 365, y= 145)
    tcard2.place(x= 545, y= 145)
    tcard3.place(x= 725, y= 145)
    tcard4.place(x= 905, y= 145)
    tcard5.place(x= 1085, y= 145)

    pcard1.place(x= 455, y= 525)
    pcard2.place(x= 635, y= 525)

        # КНОПКИ
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

    def disable(btns, btn_type):
        action[0] = btn_type
        for btn in btns:
            btn['state'] = DISABLED
            btn.bind("<Enter>", on_disable)
    
    def vote_action():
        vote[1]=True
        if vote[0]:
            vote_btn['image'] = wait_img
            vote[0]=False
            return
        vote_btn['image'] = begin_img
        vote[0]=True

    
    call_btn = Button(poker_table, 
                    image=call_img, 
                    bg=FIELD_COLOR2, 
                    activebackground=FIELD_COLOR2,
                    relief = FLAT, bd=0,
                    state=DISABLED,
                    command=lambda:disable([call_btn, 
                                            check_btn, 
                                            fold_btn, 
                                            raise1_btn, 
                                            raise2_btn, 
                                            allin_btn], 'call'))
    # call_btn.bind("<Enter>", on_enter_call)
    # call_btn.bind("<Leave>", on_leave_call)
    call_btn.place(x=850, y= 520)

    check_btn = Button(poker_table, 
                    image=check_img, 
                    bg=FIELD_COLOR2, 
                    activebackground=FIELD_COLOR2,
                    relief = FLAT, bd=0,
                    state=DISABLED,
                    command=lambda:disable([call_btn, 
                                            check_btn, 
                                            fold_btn, 
                                            raise1_btn, 
                                            raise2_btn, 
                                            allin_btn], 'check'))
                    
    # check_btn.bind("<Enter>", on_enter_check)
    # check_btn.bind("<Leave>", on_leave_check)
    check_btn.place(x=850, y= 590)

    fold_btn = Button(poker_table, 
                    image=fold_img, 
                    bg=FIELD_COLOR2, 
                    activebackground=FIELD_COLOR2,
                    relief = FLAT, bd=0,
                    state=DISABLED,
                    command=lambda:disable([call_btn, 
                                            check_btn, 
                                            fold_btn, 
                                            raise1_btn, 
                                            raise2_btn, 
                                            allin_btn], 'fold'))
    # fold_btn.bind("<Enter>", on_enter_fold)
    # fold_btn.bind("<Leave>", on_leave_fold)
    fold_btn.place(x=850, y= 660)

    raise1_btn = Button(poker_table, 
                    image=raise1_img, 
                    bg=FIELD_COLOR2, 
                    activebackground=FIELD_COLOR2,
                    relief = FLAT, bd=0,
                    state=DISABLED,
                    command=lambda:disable([call_btn, 
                                            check_btn, 
                                            fold_btn, 
                                            raise1_btn, 
                                            raise2_btn, 
                                            allin_btn], 'raise1'))
    # raise1_btn.bind("<Enter>", on_enter_raise1)
    # raise1_btn.bind("<Leave>", on_leave_raise1)
    raise1_btn.place(x=1020, y= 520)

    raise2_btn = Button(poker_table, 
                    image=raise2_img, 
                    bg=FIELD_COLOR2, 
                    activebackground=FIELD_COLOR2,
                    relief = FLAT, bd=0,
                    state=DISABLED,
                    command=lambda:disable([call_btn, 
                                            check_btn, 
                                            fold_btn, 
                                            raise1_btn, 
                                            raise2_btn, 
                                            allin_btn], 'raise2'))
    # raise2_btn.bind("<Enter>", on_enter_raise2)
    # raise2_btn.bind("<Leave>", on_leave_raise2)
    raise2_btn.place(x=1020, y= 590)

    allin_btn = Button(poker_table, 
                    image=allin_img, 
                    bg=FIELD_COLOR2, 
                    activebackground=FIELD_COLOR2,
                    relief = FLAT, bd=0,
                    state=DISABLED,
                    command=lambda:disable([call_btn, 
                                            check_btn, 
                                            fold_btn, 
                                            raise1_btn, 
                                            raise2_btn, 
                                            allin_btn], 'allin'))
    # allin_btn.bind("<Enter>", on_enter_allin)
    # allin_btn.bind("<Leave>", on_leave_allin)
    allin_btn.place(x=1020, y= 660)

    vote_btn = Button(poker_table, 
                    image=begin_img, 
                    bg=FIELD_COLOR2, 
                    activebackground=FIELD_COLOR2,
                    relief = FLAT, bd=0,
                    command=vote_action)
    vote_btn.place(x=20, y= 660)


        # ИНФОРМАЦИЯ
    bank_label = Label(poker_table, 
            anchor = 'c', 
            bg = FIELD_COLOR2, 
            fg = 'white',
            font=font_bet_big)
    bank_label.place(x = 750, y = 80)

    capital_label = Label(poker_table, 
            anchor = 'c', 
            bg = FIELD_COLOR2, 
            fg = 'white', 
            font=font_bet)
    capital_label.place(x = 860, y = 430)

    bet_label = Label(poker_table, 
            anchor = 'c', 
            bg = FIELD_COLOR2, 
            fg = 'white', 
            font=font_bet)
    bet_label.place(x = 1035, y = 430)

    combination_label = Label(poker_table, 
            anchor = 'c', 
            bg = FIELD_COLOR2, 
            fg = 'white', 
            font=font_player)
    combination_label.place(x = 470, y = 430)

    players_col = Canvas(poker_table, bg=FIELD_COLOR2, width=255, height=height-150,highlightthickness=0)
    players_col.pack(anchor=NW, expand=1)

    def show_players():
        shift = 0
        for table_player in players:
            fgc='white'
            if table_player[6]:
                fgc='gray60'
            
            players_col.create_rectangle(5, shift + 5, 250, shift + 125, outline=fgc)

            player_label = Label(poker_table, 
                    text=table_player[1],
                    anchor = 'c', 
                    bg = FIELD_COLOR2, 
                    fg = fgc, 
                    font=font_player)
            player_label.place(x = 6, y = shift + 6)

            player_label2 = Label(poker_table,
                            text='00000000', 
                            anchor = 'c', 
                            bg = FIELD_COLOR2, 
                            fg = FIELD_COLOR2, 
                            font=font_bet)
            player_label2.place(x = 6, y = shift + 85)
            
            player_label2 = Label(poker_table,
                            text=table_player[3], 
                            anchor = 'c', 
                            bg = FIELD_COLOR2, 
                            fg = fgc, 
                            font=font_bet)
            player_label2.place(x = 6, y = shift + 85)

            player_label3 = Label(poker_table,
                            text='0000000',
                            anchor = 'c', 
                            bg = FIELD_COLOR2, 
                            fg = FIELD_COLOR2, 
                            font=font_bet)
            player_label3.place(x = 135, y = shift + 85)

            player_label3 = Label(poker_table,
                            text=table_player[4], 
                            anchor = 'c', 
                            bg = FIELD_COLOR2, 
                            fg = fgc, 
                            font=font_bet)
            player_label3.place(x = 135, y = shift + 85)

            rank_img = ranks_dict[table_player[2]]

            rank_label = Label(poker_table, image=rank_img, bg=FIELD_COLOR2)
            rank_label.place(x= 10, y= shift + 50)

            if table_player[5]:

                pcard1_label = Label(poker_table, image=cover_img_mini,
                        bg = FIELD_COLOR2)
                pcard1_label.place(x = 140, y = shift + 10)

                pcard2_label = Label(poker_table, image=cover_img_mini, 
                        bg = FIELD_COLOR2)
                pcard2_label.place(x = 195, y = shift + 10)

            shift += 125
    
    def show_info():
        bank_label['text'] = info[0]
        capital_label['text'] = info[1]
        bet_label['text'] = info[2]
        combination_label['text'] = info[3]

    def get_player_info(user_uid):
        response = requests.get(BASE_URL+'/poker/v1/players/' + user_uid,
                                headers={'Authorization':TOKENS[0]})
        print(response)
        
        if response.status_code == 200:
            player_info = response.json()

            player[0] = player_info['UserUid']
            player[1] = player_info['Username']
            player[2] = player_info['UserRank']
            player[3] = '10000'
            player[4] = '0'
            player_c = player.copy()
            
            players.append(player_c)
            show_players()
    
    def winner_result(winners, combo):
        messagebox.showwarning('WINNER(S)', 'Поебедитель(-и):' + winners 
                                        + '\nКомбинация: ' + combo, parent=poker_table)
    
    def check_delay():
        sleep(3)


    
    show_players()
    show_info()

    def on_message(ws, message):
        message = json.loads(message)
        
        # delay[0] = True
        # tp=Thread(target=lambda:chek_delay()) 
        # tp.start()


        if message['MessageType'] == 'PING':
            ws.send(pong_msg)

            if auth[0]:
                auth_msg = AuthMessage(MessageId=int(time()*1000),
                        RoomUid=room_uid,
                        Token=TOKENS[0],
                        LastEventId=last_event_id
                        ).model_dump_json()
                ws.send(auth_msg)
                auth[0] = False

            if action[0] != '':
                if action[0] == 'fold':
                    act_msg = ActionMessage(MessageType='GAME-ACTION',
                            MessageId=int(time()*1000),
                            RoomUid=room_uid,
                            UserUid=TOKENS[2],
                            ActionType='FOLD'
                            ).model_dump_json()
                    ws.send(act_msg)
                    action[0] = ''
                
                elif action[0] == 'check':
                    act_msg = ActionMessage(MessageType='GAME-ACTION',
                            MessageId=int(time()*1000),
                            RoomUid=room_uid,
                            UserUid=TOKENS[2],
                            ActionType='CHECK'
                            ).model_dump_json()
                    ws.send(act_msg)
                    action[0] = ''
                
                elif action[0] == 'call':
                    act_msg = ActionMessage(MessageType='GAME-ACTION',
                            MessageId=int(time()*1000),
                            RoomUid=room_uid,
                            UserUid=TOKENS[2],
                            ActionType='CALL'
                            ).model_dump_json()
                    ws.send(act_msg)
                    action[0] = ''
                
                elif action[0] == 'raise1':
                    act_msg = ActionMessage(MessageType='GAME-ACTION',
                            MessageId=int(time()*1000),
                            RoomUid=room_uid,
                            UserUid=TOKENS[2],
                            ActionType='RAISE',
                            Coef='X1_5'
                            ).model_dump_json()
                    ws.send(act_msg)
                    action[0] = ''
                
                elif action[0] == 'raise2':
                    act_msg = ActionMessage(MessageType='GAME-ACTION',
                            MessageId=int(time()*1000),
                            RoomUid=room_uid,
                            UserUid=TOKENS[2],
                            ActionType='RAISE',
                            Coef='X2'
                            ).model_dump_json()
                    ws.send(act_msg)
                    action[0] = ''
                
                elif action[0] == 'allin':
                    act_msg = ActionMessage(MessageType='GAME-ACTION',
                            MessageId=int(time()*1000),
                            RoomUid=room_uid,
                            UserUid=TOKENS[2],
                            ActionType='RAISE',
                            Coef='ALL-IN'
                            ).model_dump_json()
                    ws.send(act_msg)
                    action[0] = ''
            
            if vote[1]:
                vote[1] = False
                if vote[0]:
                    act_msg = ActionMessage(MessageType='VOTE',
                        MessageId=int(time()*1000),
                        RoomUid=room_uid,
                        UserUid=TOKENS[2],
                        VoteType='WAIT'
                        ).model_dump_json()
                    ws.send(act_msg)
                else:
                    act_msg = ActionMessage(MessageType='VOTE',
                        MessageId=int(time()*1000),
                        RoomUid=room_uid,
                        UserUid=TOKENS[2],
                        VoteType='START'
                        ).model_dump_json()
                    ws.send(act_msg)


        
        elif message['MessageType'] == 'EVENT':
    
            if message['EventType'] == 'PLAYER-ACTION-EVENT':
                
                user_uid = message['EventDescriptor']['UserUid']
                
                if message['EventDescriptor']['ActionType'] == 'INCOME':
                    f = False
                    for table_player in players:
                        if user_uid != table_player[0]:
                            f = True

                    if f:
                        tp=Thread(target=lambda:get_player_info(user_uid)) 
                        tp.start()
            
                elif message['EventDescriptor']['ActionType'] == 'OUTCOME':
                    for table_player in players:
                        if user_uid == table_player[0]:
                            table_player[6] = True
                            show_players()
                    if user_uid == TOKENS[2]:
                        messagebox.showerror('YOU LOOSE', 'К сожалению, вы проиграли!', parent=poker_table)
                        # poker_table.destroy()

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
                                
                                for raise_variant in bout_variant['RaiseVariants']:
                                    if raise_variant == 'X1_5':
                                        raise1_btn.config(state=NORMAL)
                                        raise1_btn.bind("<Enter>", on_enter_raise1)
                                        raise1_btn.bind("<Leave>", on_leave_raise1)
                                    elif raise_variant == 'X2':
                                        raise2_btn.config(state=NORMAL)
                                        raise2_btn.bind("<Enter>", on_enter_raise2)
                                        raise2_btn.bind("<Leave>", on_leave_raise2)
                                    elif raise_variant == 'ALL-IN':
                                        allin_btn.config(state=NORMAL)
                                        allin_btn.bind("<Enter>", on_enter_allin)
                                        allin_btn.bind("<Leave>", on_leave_allin)
                        
                        show_info()
                
                elif message['EventDescriptor']['ActionType'] == 'FOLD':
                    for table_player in players:
                        if table_player[0] == user_uid:
                            table_player[6] = True
                    show_players()
                    if user_uid == TOKENS[2]:
                        disable([call_btn, check_btn, fold_btn,
                                 raise1_btn, raise2_btn, allin_btn], '')

                
                elif message['EventDescriptor']['ActionType'] == 'CHECK':
                    if user_uid == TOKENS[2]:
                        disable([call_btn, check_btn, fold_btn,
                                 raise1_btn, raise2_btn, allin_btn], '')
                
                elif message['EventDescriptor']['ActionType'] == 'CALL':
                    for table_player in players:
                        if table_player[0] == user_uid:
                            table_player[3] = str(message['EventDescriptor']['NewDeposit'])
                            table_player[4] = str(message['EventDescriptor']['NewBet'])
                    if user_uid == TOKENS[2]:
                        info[1] = str(message['EventDescriptor']['NewDeposit'])
                        info[2] = str(message['EventDescriptor']['NewBet'])
                        disable([call_btn, check_btn, fold_btn,
                                 raise1_btn, raise2_btn, allin_btn], '')
                    show_players()
                    show_info()
                
                elif message['EventDescriptor']['ActionType'] == 'RAISE':
                    for table_player in players:
                        if table_player[0] == user_uid:
                            table_player[3] = str(message['EventDescriptor']['NewDeposit'])
                            table_player[4] = str(message['EventDescriptor']['NewBet'])
                    if user_uid == TOKENS[2]:
                        info[1] = str(message['EventDescriptor']['NewDeposit'])
                        info[2] = str(message['EventDescriptor']['NewBet'])
                        disable([call_btn, check_btn, fold_btn,
                                 raise1_btn, raise2_btn, allin_btn], '')
                    show_players()
                    show_info()
                
                elif message['EventDescriptor']['ActionType'] == 'ALL-IN':
                    for table_player in players:
                        if table_player[0] == user_uid:
                            table_player[3] = str(message['EventDescriptor']['NewDeposit'])
                            table_player[4] = str(message['EventDescriptor']['NewBet'])
                    if user_uid == TOKENS[2]:
                        info[1] = str(message['EventDescriptor']['NewDeposit'])
                        info[2] = str(message['EventDescriptor']['NewBet'])
                        disable([call_btn, check_btn, fold_btn,
                                 raise1_btn, raise2_btn, allin_btn], '')
                    show_players()
                    show_info()
                
                elif message['EventDescriptor']['ActionType'] == 'SET-DEALER':
                    pass
                
                elif message['EventDescriptor']['ActionType'] == 'MIN-BLIND-IN':
                    print()
                    for table_player in players:
                        print(table_player)
                        if table_player[0] == user_uid:
                            table_player[3] = str(message['EventDescriptor']['NewDeposit'])
                            table_player[4] = str(message['EventDescriptor']['NewBet'])
                            print(table_player)
                    print()
                    if user_uid == TOKENS[2]:
                        info[1] = str(message['EventDescriptor']['NewDeposit'])
                        info[2] = str(message['EventDescriptor']['NewBet'])

                    show_players()
                    show_info()
                
                elif message['EventDescriptor']['ActionType'] == 'MAX-BLIND-IN':
                    for table_player in players:
                        if table_player[0] == user_uid:
                            table_player[3] = str(message['EventDescriptor']['NewDeposit'])
                            table_player[4] = str(message['EventDescriptor']['NewBet'])
                    if user_uid == TOKENS[2]:
                        info[1] = str(message['EventDescriptor']['NewDeposit'])
                        info[2] = str(message['EventDescriptor']['NewBet'])
                    show_players()
                    show_info()
    
            elif message['EventType'] == 'GAME-EVENT':
            
                if message['EventDescriptor']['EventType'] == 'ROOM_STATE_UPDATE':
                    if message['EventDescriptor']['NewRoomState'] == 'GAMING':
                        vote_btn.destroy()
                        # tcard1.config(image=cover_img)
                        # tcard2.config(image=cover_img)
                        # tcard3.config(image=cover_img)
                        # tcard4.config(image=cover_img)
                        # tcard5.config(image=cover_img)
                        # pcard1.config(image=cover_img)
                        # pcard2.config(image=cover_img)
                    elif message['EventDescriptor']['NewRoomState'] == 'DISSLOLUTION':
                        messagebox.showerror('YOU WIN', 'Поздравляем, вы победили!')
                        ws.close()
                        poker_table.destroy()
                
                elif message['EventDescriptor']['EventType'] == 'NEW_ROUND':
                    tcard1.config(image=cover_img)
                    tcard2.config(image=cover_img)
                    tcard3.config(image=cover_img)
                    tcard4.config(image=cover_img)
                    tcard5.config(image=cover_img)
                    pcard1.config(image=cover_img)
                    pcard2.config(image=cover_img)
                    for table_player in players:
                        table_player[5] = True
                        table_player[6] = False
                
                elif message['EventDescriptor']['EventType'] == 'NEW_TRADE_ROUND':
                    pass
                
                elif message['EventDescriptor']['EventType'] == 'PERSONAL_CARDS':
                    suit1 = message['EventDescriptor']['PlayingCardsList'][0]['CardSuit']
                    index1 = message['EventDescriptor']['PlayingCardsList'][0]['Index']
                    suit2 = message['EventDescriptor']['PlayingCardsList'][1]['CardSuit']
                    index2 = message['EventDescriptor']['PlayingCardsList'][1]['Index']

                    pcard1.config(image=cards_dict[suit1][index1])
                    pcard2.config(image=cards_dict[suit2][index2])

                    info[3] = message['EventDescriptor']['BestCombName']
                    show_info()
                
                
                elif message['EventDescriptor']['EventType'] == 'CARDS_ON_TABLE':
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
                    
                    info[3] = message['EventDescriptor']['BestCombName']
                    show_info()
                
                
                elif message['EventDescriptor']['EventType'] == 'BET_ACCEPTED':
                    for table_player in players:
                        table_player[4] = ''
                    info[0] = str(message['EventDescriptor']['NewStack'])
                    show_players()
                    show_info()
                
                
                elif message['EventDescriptor']['EventType'] == 'WINNER_RESULT':
                    winners = ''
                    i = 0
                    for uuid in message['EventDescriptor']['WinnerUids']:
                        for table_player in players:
                            if uuid == table_player[0]:
                                table_player[3] == str(message['EventDescriptor']['WinnerDeposits'][i])
                                i+=1
                                winners+= ' ' + table_player[1]
                    combo = message['EventDescriptor']['BestCombName']


                    info[0] = '0'          
                    
                    show_players()
                    show_info()

                    tp=Thread(target=lambda:winner_result(winners, combo)) 
                    tp.start()
    
    def on_open(ws):
        print("##### opened #####")

    websocket.enableTrace(True)
    ws = websocket.WebSocketApp(WS_BASE_URL 
                                + 'poker/v1/rooms-ws/'
                                + room_uid + '?uid=' +  TOKENS[2],
                                on_message = on_message,
                                on_open= on_open)
    wst = Thread(target=ws.run_forever)
    wst.daemon = True
    wst.start()
    
    poker_table.mainloop()

# poker_table()