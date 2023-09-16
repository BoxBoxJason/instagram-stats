# -*- coding: utf-8 -*-
'''
Created on 4 mars 2023

@author: BoxBoxJason
'''

import os
from matplotlib import pyplot
from Resources import enums
import logging


resultDirPath = os.getenv("RESULTSDIRPATH")
pyplot.rcParams["figure.figsize"] = [7.00, 3.50]
pyplot.rcParams["figure.autolayout"] = True


def textRanking(ranking,total,rankingType,title,fileName):
    """
    Creates a text ranking file with requested input type
    """
    logging.debug(f"Creating {title} ranking file")
    textResultFilePath = os.path.join(resultDirPath,f"{fileName}.txt")
    with open(textResultFilePath,'w',encoding="utf-8") as textResultFile:
        textResultFile.write(title.format(total) + '\n')
        ranking = dict(sorted(ranking.items(),reverse = True))
        for indexUser,(countUser,user) in enumerate(ranking.items()):
            indexUser += 1
            text = f"{indexUser}: {countUser} {rankingType}: {','.join(user)}\n"
            textResultFile.write(text)


def pieChartFigure(countList,usersList,titleFigure,titreGraph,colors,total = None):
    """
    Constructs a pie chart with provided data
    """
    logging.debug(f"Creating {titleFigure} pie chart")
    if total is None:
        total = sum(countList)
    pyplot.figure(titleFigure,facecolor='black')
    pyplot.title(f"{titreGraph} (Total {total})",fontsize=15, color= 'white', fontweight='bold')
    _,labeltext,pctg=pyplot.pie(countList,autopct='%1.1f%%',labels=usersList,colors = colors)
    for i,labelText in enumerate(labeltext):
        labelText.set_color('white')
        pctg[i].set_color('white')
    manager = pyplot.get_current_fig_manager()
    manager.full_screen_toggle()
    resultPath = os.path.join(resultDirPath,titleFigure)
    pyplot.savefig(resultPath)


def heatMapFigure(grid,titleFigure,titleGraph,labelX,labelY):
    """
    Constructs a pie chart with provided data
    """
    logging.debug(f"Creating {titleFigure} heatmap")
    pyplot.figure(titleFigure,facecolor='black')
    pyplot.imshow(grid,cmap = 'coolwarm',interpolation='nearest',aspect='auto')
    pyplot.title(titleGraph,fontsize=15, color= 'white', fontweight='bold')
    pyplot.xticks(range(7),enums.DAYS,color='white')
    pyplot.xlabel(labelX,fontsize=12, color= 'gray', fontweight='bold')
    pyplot.yticks(range(24),range(24),color='white')
    pyplot.ylabel(labelY,fontsize=12, color= 'gray', fontweight='bold',rotation='vertical')
    colorbar = pyplot.colorbar()
    colorbarax = pyplot.getp(colorbar.ax.axes, 'yticklabels')
    pyplot.setp(colorbarax, color='white')
    colorbar.colors='white'
    manager = pyplot.get_current_fig_manager()
    manager.full_screen_toggle()
    resultPath = os.path.join(resultDirPath,titleFigure)
    pyplot.savefig(resultPath)


def barGraphFigure(xdata,ydata,titleFigure,titleGraph,labelX,labelY,edgeColor=None):
    """
    Constructs a bar graph with provided data
    """
    logging.debug(f"Creating {titleFigure} bar graph")
    pyplot.figure(titleFigure,facecolor = 'black')
    pyplot.title(titleGraph,fontsize=15, color= 'white', fontweight='bold')
    if edgeColor is None:
        pyplot.bar(xdata,ydata,width=3,color='#000359')
    else:
        pyplot.bar(xdata,ydata,color='blue',edgecolor=edgeColor)
    pyplot.xticks(rotation='vertical',color='white')
    pyplot.xlabel(labelX,fontsize=12, color= 'gray', fontweight='bold')
    pyplot.yticks(color='white')
    pyplot.ylabel(labelY,fontsize=12, color= 'gray', fontweight='bold',rotation='vertical')
    manager = pyplot.get_current_fig_manager()
    manager.full_screen_toggle()
    resultPath = os.path.join(resultDirPath,titleFigure)
    pyplot.savefig(resultPath)
