# -*- coding: utf-8 -*-
'''
Created on 15 sept. 2023

@author: BoxBoxJason
'''
from datetime import datetime,date
import json
from json.decoder import JSONDecodeError
import logging
import os
import re

from Resources.enums import USERDICT


def collectData(sourceFolderPath,wordsToSearchFor):
    """
    Explores all files and collects valuable data from it
    """
    # List of files in source folder
    FILES = os.listdir(sourceFolderPath)
    # Number of files in source folder
    NUMBEROFFILES = len(FILES)
    # Users data dict
    usersDict = {}
    # Count of messages per day
    daysCount = {}
    # Count of messages per hour and day of week
    gridHourDay = [[0 for i in range(7)] for j in range(24)]

    #Beginning of files data collection
    for fileIndex,jsonFileName in enumerate(FILES):
        FILEOK = True
        fileIndex += 1
        logging.debug(f"Collecting data from {jsonFileName} (file {fileIndex}/{NUMBEROFFILES})")

        jsonFilePath = os.path.join(sourceFolderPath,jsonFileName)
        if jsonFilePath.endswith(".json") and os.path.exists(jsonFilePath) and os.access(jsonFilePath,os.R_OK):
            with open(jsonFilePath,'r',encoding="utf-8") as openedJsonFile:
                try:
                    jsonMessagesData = json.load(openedJsonFile)
                except (TypeError,JSONDecodeError):
                    FILEOK = False
        else:
            logging.error(f"{jsonFilePath} could not be opened or read as a .json file")
            FILEOK = False

        if FILEOK:
            #Creating users dictionary with all retrieved data
            participants = jsonMessagesData.get("participants",[])
            for participant in participants:
                if "name" in participant and not participant["name"] in usersDict:
                    usersDict[participant["name"]] = USERDICT(participant["name"])

            #Collecting messages data
            for message in jsonMessagesData["messages"]:
                collectMessageData(usersDict, daysCount, gridHourDay, wordsToSearchFor, message)

    return usersDict,daysCount,gridHourDay


def collectMessageData(usersDict,daysCount,gridHourDay,wordsToSearchFor,message):
    """
    Collects message data and stores it into corresponding dict, list...
    """
    user = message["sender_name"]
    mes = message.get("content")
    if mes != "Liked a message":
        #Determining type of message
        if "share" in message: #reel
            usersDict[user]["reels"] += 1
        elif "photos" in message: #photo
            usersDict[user]["photosAndVideos"] += 1
        elif "videos" in message: #video
            usersDict[user]["photosAndVideos"] += 1
        else:
            if mes is not None:
                #Searching for particular words / phrases given in enums
                for wordToSearch in wordsToSearchFor:
                    if wordToSearch in mes.lower():
                        usersDict[user]["wordsToFind"] += 1
                #Vocabulary analysis
                mesSplit = re.sub(r' \W+', '', mes).lower().split(' ')
                for word in mesSplit:
                    usersDict[user]["vocabulary"].add(word)
            usersDict[user]["messages"] += 1

        #Creating time statistics
        timestampmsg = message["timestamp_ms"]/1000
        msgdate = datetime.fromtimestamp(timestampmsg)
        day = date.fromtimestamp(timestampmsg)
        if day in daysCount:
            daysCount[day] += 1
        else:
            daysCount[day] = 1
        dayOfWeek = msgdate.isoweekday() - 1
        hour = msgdate.hour
        gridHourDay[hour][dayOfWeek] += 1
        reactions = message.get("reactions")
        if reactions is not None:
            for reaction in reactions:
                usersDict[reaction["actor"]]["reactions"] += 1
