# Test to find n'th fibonacci number

	.data
nStr:	.asciiz "Enter n: "
n:	.word	0
i:	.word	0
retVal:	.word	0
str:	.asciiz "n'th fibonacci number: "
first:	.word	0
second:	.word	0
temp:	.word	0

	.text


	.globl main
	.ent main
main:
	li $v0, 4
	la $a0, nStr
	syscall
	li $v0, 5
	syscall
	move $t1, $v0
	li $t4, 0		# i -> $t4
	sw $t1, n
	sw $t4, i
	jal fib
	lw $t1, n
	lw $t4, i
	sw $t1, n		# spilled n, freed $t1
	move $t1, $v0
	li $v0, 4
	la $a0, str
	syscall
	li $v0, 1
	move $a0, $t1
	syscall
	# Store variables back into memory
	sw $t1, retVal
	sw $t4, i
	li $v0, 10
	syscall
	.end main

	.globl fib
	.ent fib
fib:
	li $t1, 1		# first -> $t1
	li $t4, 1		# second -> $t4
	sw $t1, first		# spilled first, freed $t1
	lw $t1, n
	sub $t1, $t1, 2		# n -> $t1
	# Store variables back into memory
	sw $t1, n
	sw $t4, second

loop:

	lw $t1, n
	ble $t1, 0, exit		# exit -> $t0
	lw $t4, second
	move $t3, $t4		# temp -> $t3
	lw $t2, first
	add $t4, $t4, $t2	# second -> $t4
	move $t2, $t3		# first -> $t2
	sub $t1, $t1, 1		# n -> $t1
	# Store variables back into memory
	sw $t1, n
	sw $t4, second
	sw $t3, temp
	sw $t2, first

	j loop

exit:
	lw $v0, second

	jr $ra
	.end fib
