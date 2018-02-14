# Maximum XOR-value of at-most k-elements from 1 to n

	.data
n:	.word	0
k:	.word	0
x:	.word	0
result:	.word	0

	.text


	.globl main
	.ent main
main:
	li $v0, 5
	syscall
	move $t1, $v0
	li $v0, 5
	syscall
	move $t4, $v0
	# x = log2(n) + 1
	sw $t1, n		# spilled n, freed $t1
	li $t1, 0		# x -> $t1

	# Store variables back into memory
	sw $t1, x
	sw $t4, k

while:
	lw $t1, n
	srl $t1, $t1, 1		# n -> $t1
	lw $t4, x
	addi $t4, $t4, 1	# x -> $t4

	# Store variables back into memory
	sw $t1, n
	sw $t4, x

	lw $t1, n
	bgt $t1, 0, while		# while -> $t0
	# Return (2^x - 1)
	li $t4, 1		# result -> $t4
	sw $t1, n		# spilled n, freed $t1
	lw $t1, x
	sll $t4, $t4, $t1	# result -> $t4
	sub $t4, $t4, 1		# result -> $t4
	li $v0, 1
	move $a0, $t4
	syscall

	# Store variables back into memory
	sw $t1, x
	sw $t4, result
	li $v0, 10
	syscall
	.end main
