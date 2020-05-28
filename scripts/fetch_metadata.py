# A script for fetching metadata from video file

import cv2
import argparse
import os.path
import json
import logging

parser = argparse.ArgumentParser()
parser.add_argument("-i","--input", help="input video path", required=True)
parser.add_argument("-o","--output", help='output file path')
args = parser.parse_args()

file_path = args.input
video_name = os.path.splitext(file_path)[0] #remove extension from file

if args.output:
    metadata_filename = args.output
else:
    metadata_filename = "{0}_metadata.txt".format(video_name)

logging.debug("file path: %s", file_path)
logging.debug("file name: %s", video_name)


if not os.path.exists(file_path) or not os.path.isfile(file_path): #checks for existance of file
    raise Exception("can't read file {0}".format(file_path))

video = cv2.VideoCapture(file_path)

height = video.get(cv2.CAP_PROP_FRAME_HEIGHT)
width  = video.get(cv2.CAP_PROP_FRAME_WIDTH) 
frames_count = video.get(cv2.CAP_PROP_FRAME_COUNT ) 
fps = video.get(cv2.CAP_PROP_FPS)

logging.info("fetched metadata from file %s",file_path)

duration = 0

if fps != 0:
    duration = frames_count/fps
else:
    logging.warning("file %s has zero fps, or empty file"%file_path)

metadata = {
    "height":int(height),
    "width":int(width),
    "framesCount":int(frames_count),
    "fps":fps,
    "duration":duration
}

print(metadata)

logging.debug("writing metadata to file %s",metadata_filename)

with open(metadata_filename, 'w') as outfile:
    #print metadata to file
    json.dump(metadata, outfile)
