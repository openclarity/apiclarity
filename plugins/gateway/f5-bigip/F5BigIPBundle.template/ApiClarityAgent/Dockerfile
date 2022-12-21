FROM python:3.8.7-slim-buster

# Copy local code to the container image.
ENV APP_HOME /app
WORKDIR $APP_HOME

COPY ./requirements.txt requirements.txt

RUN pip3 install -r requirements.txt
COPY src ./

EXPOSE 10514/udp

ENTRYPOINT ["python3", "APIClarityAgent.py"]