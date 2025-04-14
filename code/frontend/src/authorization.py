from tkinter import *
from tkinter import messagebox
from tkinter import font
from PIL import ImageTk, Image
import requests

from color import FIELD_COLOR, FONT_COLOR

from poker import poker_table

from schemas.request import AuthenticationRequest, SignUpRequest
from config import BASE_URL, TOKENS

def pre_authorize(root, login, password, password2):

    if password != password2:
        messagebox.showerror('Password Error', 'Введенные пароли не совпадают!')
        return
    
    signup_request = SignUpRequest(scope='OPENID', username=login, password=password).model_dump_json()

    print("\nЗапрос регистрации")
    print(signup_request)

    signup_response = requests.post(BASE_URL+'/poker/v1/register', signup_request)

    print("\nОтвет на запрос регистрации")
    print(signup_response.status_code)
    print(signup_response.json())

    if signup_response.status_code == 400:
        messagebox.showerror('Error 400', 'Аккаунт с указанным логином уже существует!')
        return
    elif signup_response.status_code == 500:
        messagebox.showerror('Error 500', 'Внутренняя ошибка сервера!')
        return
    
    play(root, login, password)


def my_account(root):

    print("\nЗапрос информации о пользователе...")

    user_info_response = requests.get(BASE_URL+'/poker/v1/me', 
                                      headers={'Authorization':TOKENS[0]})
    
    print("\nОтвет на запрос о получении данных авторизованного пользователя")
    print(user_info_response.status_code)
    print(user_info_response.json())

    if user_info_response.status_code == 401:
        messagebox.showerror('Error 401', 'Время действия AccessToken истекло!')
        return
    elif user_info_response.status_code == 500:
        messagebox.showerror('Error 500', 'Внутренняя ошибка сервера!')
        return 
    
    user_info = user_info_response.json()

    # root.iconify()
    
    account = Toplevel()
    account.grab_set()

    width = account.winfo_screenwidth()
    height = account.winfo_screenheight()
    x = (width - 750) / 2
    y = (height - 440) / 2

    account.geometry('750x440+%d+%d' % (x, y))
    account.title('TEXAS HOLDEM')
    account.resizable(False, False)
    account.configure(background = FIELD_COLOR)
    account.iconbitmap("icon.ico")

    font_reg = font.Font(family="Century Gothic", size=14)
    font_reg_big = font.Font(family="Century Gothic", size=25, weight='bold')

    poker_screen = ImageTk.PhotoImage(Image.open("entry_screen.jpg"))
    screen = Label(account, image = poker_screen, bg = FIELD_COLOR)
    screen.image_ref = poker_screen
    screen.pack()
    screen.place(x = -5, y = -5)

    play_label = Label(account, text='TEXAS HOLDEM', 
                    anchor = 'c', 
                    bg = FIELD_COLOR, 
                    fg = FONT_COLOR, 
                    font=font_reg_big)
    play_label.place(x = 445, y = 40)

    player_label = Label(account, text='Имя:', 
                    anchor = 'c', 
                    bg = FIELD_COLOR, 
                    fg = 'white', 
                    font=font_reg)
    player_label.place(x = 410, y = 110)

    player_label = Label(account, text='Игр сыграно:', 
                    anchor = 'n', 
                    bg = FIELD_COLOR, 
                    fg = 'white', 
                    font=font_reg)
    player_label.place(x = 410, y = 160)

    player_label = Label(account, text='Всего побед:', 
                    anchor = 'n', 
                    bg = FIELD_COLOR, 
                    fg = 'white', 
                    font=font_reg)
    player_label.place(x = 410, y = 210)

    player_label = Label(account, text='Винрейт:', 
                    anchor = 'n', 
                    bg = FIELD_COLOR, 
                    fg = 'white', 
                    font=font_reg)
    player_label.place(x = 410, y = 260)

    player_label = Label(account, text='Ранг:', 
                    anchor = 'c', 
                    bg = FIELD_COLOR, 
                    fg = 'white', 
                    font=font_reg)
    player_label.place(x = 410, y = 310)

    player_label2 = Label(account, text=user_info['Username'], 
                    anchor = 'c', 
                    bg = FIELD_COLOR, 
                    fg = FONT_COLOR, 
                    font=font_reg_big)
    player_label2.place(x = 550, y = 110)

    player_label2 = Label(account, text=str(user_info['NumOfGames']), 
                    anchor = 'c', 
                    bg = FIELD_COLOR, 
                    fg = FONT_COLOR, 
                    font=font_reg_big)
    player_label2.place(x = 550, y = 160)

    player_label2 = Label(account, text=str(user_info['NumOfWins']), 
                    anchor = 'c', 
                    bg = FIELD_COLOR, 
                    fg = FONT_COLOR, 
                    font=font_reg_big)
    player_label2.place(x = 550, y = 210)

    if user_info['NumOfGames'] == 0:
        text = '0'
    else:
        text=str(int(user_info['NumOfWins']/user_info['NumOfGames']*100))

    player_label2 = Label(account, text=text + '%', 
                    anchor = 'c', 
                    bg = FIELD_COLOR, 
                    fg = FONT_COLOR, 
                    font=font_reg_big)
    player_label2.place(x = 550, y = 260)

    player_label2 = Label(account, text=str(user_info['UserRank']), 
                    anchor = 'c', 
                    bg = FIELD_COLOR, 
                    fg = FONT_COLOR, 
                    font=font_reg_big)
    player_label2.place(x = 550, y = 310)


def play(root, login, password):

    authorization_request = AuthenticationRequest(scope='OPENID', 
                                                 grantType='PASSWORD', 
                                                 username=login, 
                                                 password=password).model_dump_json()
    
    print("\n Запрос на авторизацию пользователя")
    print(authorization_request)
    
    authorization_response = requests.post(BASE_URL + '/poker/v1/oauth/token', 
                                          authorization_request)
    
    print("\n Ответ на запрос авторизации пользователя")
    print(authorization_response.status_code)
    print(authorization_response.json())
    
    if authorization_response.status_code == 401:
        messagebox.showerror('Error 401', 'Неверное имя пользователя или пароль!')
        return
    elif authorization_response.status_code == 500:
        messagebox.showerror('Error 500', 'Внутренняя ошибка сервера!')
        return
    
    accessToken = authorization_response.json()['AccessToken']
    refreshToken = authorization_response.json()['RefreshToken']
    userUid = authorization_response.json()['UserUid']

    TOKENS[0] = accessToken
    TOKENS[1] = refreshToken
    TOKENS[2] = userUid
    TOKENS[3] = login

    root.destroy()
    
    play = Tk()
    play.grab_set()
    
    width = play.winfo_screenwidth()
    height = play.winfo_screenheight()
    x = (width - 750) / 2
    y = (height - 440) / 2

    play.geometry('750x440+%d+%d' % (x, y))
    play.title('TEXAS HOLDEM')
    play.resizable(False, False)
    play.configure(background = FIELD_COLOR)
    play.iconbitmap("icon.ico")

    font_reg = font.Font(family="Century Gothic", size=14)
    font_reg_big = font.Font(family="Century Gothic", size=25, weight='bold')

    poker_screen = ImageTk.PhotoImage(Image.open("entry_screen.jpg"))
    screen = Label(play, image = poker_screen, bg = FIELD_COLOR)
    screen.image_ref = poker_screen
    screen.pack()
    screen.place(x = -5, y = -5)


    play_label = Label(play, text='TEXAS HOLDEM', 
                    anchor = 'c', 
                    bg = FIELD_COLOR, 
                    fg = FONT_COLOR, 
                    font=font_reg_big)
    play_label.place(x = 445, y = 40)

    play_btn = Button(play, text='Играть',
                        width=20,
                        height=1,
                        font=font_reg,
                        bg = FIELD_COLOR,
                        fg = FONT_COLOR,
                        relief = RIDGE,
                        command=poker_table)
    play_btn.place(anchor = 'w', x = 450, y = 150)

    me_btn = Button(play, text='Мой аккаунт',
                        width=20,
                        height=1,
                        font=font_reg,
                        bg = FIELD_COLOR,
                        fg = FONT_COLOR,
                        relief = RIDGE,
                        command=lambda:my_account(play))
    me_btn.place(anchor = 'w', x = 450, y = 225)

    info_btn = Button(play, text='Комбинации и ранги',
                        width=20,
                        height=1,
                        font=font_reg,
                        bg = FIELD_COLOR,
                        fg = FONT_COLOR,
                        relief = RIDGE)
    info_btn.place(anchor = 'w', x = 450, y = 300)


    

def registration(root):
    root.destroy()
    
    reg = Tk()
    reg.grab_set()
    
    width = reg.winfo_screenwidth()
    height = reg.winfo_screenheight()
    x = (width - 750) / 2
    y = (height - 440) / 2

    reg.geometry('750x440+%d+%d' % (x, y))
    reg.title('TEXAS HOLDEM')
    reg.resizable(False, False)
    reg.configure(background = FIELD_COLOR)
    reg.iconbitmap("icon.ico")

    font_reg = font.Font(family="Century Gothic", size=14)
    font_reg_big = font.Font(family="Century Gothic", size=25, weight='bold')

    poker_screen = ImageTk.PhotoImage(Image.open("entry_screen.jpg"))
    screen = Label(reg, image = poker_screen, bg = FIELD_COLOR)
    screen.image_ref = poker_screen
    screen.pack()
    screen.place(x = -5, y = -5)

    reg_label = Label(reg, text='Регистрация', 
                    anchor = 'c', 
                    bg = FIELD_COLOR, 
                    fg = FONT_COLOR, 
                    font=font_reg_big)
    reg_label.place(x = 450, y = 40)

    login_label = Label(reg, text='Логин', 
                    anchor = 'w', 
                    bg = FIELD_COLOR, 
                    fg = FONT_COLOR, 
                    font=font_reg)
    login_label.place(x = 450, y = 100)

    login_entry = Entry(reg, width = 20, font = font_reg) 
    login_entry.place(x = 452, y = 130)

    password_label = Label(reg, text='Пароль', 
                    anchor = 'w', 
                    bg = FIELD_COLOR, 
                    fg = FONT_COLOR, 
                    font=font_reg)
    password_label.place(x = 450, y = 160)

    password_entry = Entry(reg, width = 20, show = '*', font = font_reg) 
    password_entry.place(x = 452, y = 190)
    
    password_label2 = Label(reg, text='Подтвердите пароль', 
                    anchor = 'w', 
                    bg = FIELD_COLOR, 
                    fg = FONT_COLOR, 
                    font=font_reg)
    password_label2.place(x = 450, y = 220)

    password_entry2 = Entry(reg, width = 20, show = '*', font = font_reg) 
    password_entry2.place(x = 452, y = 250)

    reg_btn = Button(reg, text='Зарегистрироваться',
                        width=20,
                        height=1,
                        font=font_reg,
                        bg = FIELD_COLOR,
                        fg = FONT_COLOR,
                        relief = RIDGE,
                        command=lambda:pre_authorize(reg, login_entry.get(), 
                                            password_entry.get(), 
                                            password_entry2.get()))
    reg_btn.place(anchor = 'w', x = 450, y = 340)
    

def authorization():
    root = Tk()

    width = root.winfo_screenwidth()
    height = root.winfo_screenheight()
    x = (width - 750) / 2
    y = (height - 440) / 2

    root.geometry('750x440+%d+%d' % (x, y))
    root.title('TEXAS HOLDEM')
    root.resizable(False, False)
    root.configure(background = FIELD_COLOR)
    root.iconbitmap("icon.ico")

    font_reg = font.Font(family="Century Gothic", size=14)
    font_reg_big = font.Font(family="Century Gothic", size=25, weight='bold')

    poker_screen = ImageTk.PhotoImage(Image.open("entry_screen.jpg"))
    screen = Label(root, image = poker_screen, bg = FIELD_COLOR)
    screen.image_ref = poker_screen
    screen.pack()
    screen.place(x = -5, y = -5)

    enter_label = Label(text='Авторизация', 
                    anchor = 'c', 
                    bg = FIELD_COLOR, 
                    fg = FONT_COLOR, 
                    font=font_reg_big)
    enter_label.place(x = 450, y = 40)

    login_label = Label(text='Логин', 
                    anchor = 'w', 
                    bg = FIELD_COLOR, 
                    fg = FONT_COLOR, 
                    font=font_reg)
    login_label.place(x = 450, y = 100)

    login_entry = Entry(root, width = 20, font = font_reg) 
    login_entry.place(x = 452, y = 130)

    password_label = Label(text='Пароль', 
                    anchor = 'w', 
                    bg = FIELD_COLOR, 
                    fg = FONT_COLOR, 
                    font=font_reg)
    password_label.place(x = 450, y = 160)

    password_entry = Entry(root, width = 20, show = '*', font = font_reg) 
    password_entry.place(x = 452, y = 190)

    login_btn = Button(text='Войти',
                        width=20,
                        height=1,
                        font=font_reg,
                        bg = FIELD_COLOR,
                        fg = FONT_COLOR,
                        relief = RIDGE,
                        command=lambda:play(root, login_entry.get(), password_entry.get()))
    login_btn.place(anchor = 'w', x = 450, y = 260)

    reg_label = Label(text='Нет аккаунта?', 
                    anchor = 'w', 
                    bg = FIELD_COLOR, 
                    fg = FONT_COLOR, 
                    font=font_reg)
    reg_label.place(x = 450, y = 300)

    reg_btn = Button(text='Зарегистрироваться',
                        width=20,
                        height=1,
                        font=font_reg,
                        bg = FIELD_COLOR,
                        fg = FONT_COLOR,
                        relief = RIDGE,
                        command=lambda:registration(root))
    reg_btn.place(anchor = 'w', x = 450, y = 350)

    root.mainloop()
    
authorization()